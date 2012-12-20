// Copyright 2012 polaris(studygolang.com). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// project包负责管理需要自动编译、运行的项目（Project）
package project

import (
    "bytes"
    "errors"
    "files"
    "fsnotify"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "text/template"
    "time"
)

const pathSeparator = string(os.PathSeparator)

var (
    errorTplFile = "templates/error.html"

    tpl *template.Template

    successFlag = "finished"

    PrjRootErr = errors.New("project can't be found'!")
    // 项目编译有语法错误
    PrjSyntaxError = errors.New("project source syntax error!")
)

func init() {
    tpl = template.Must(template.ParseFiles(errorTplFile))
}

// Watch 监听项目
// 
// name：项目名称（最后生成的可执行程序名，不包括后缀）；
// root: 项目根目录
// goWay: 编译项目的方式，run、build还是install
// deamon: 项目是否是一直运行的（即不手动退出，程序不会终止，一般会有死循环，比如Web服务）
// mainFile：main包的main函数所在文件路径（相对于src目录）
// depends：是依赖的其他GOPATH路径下的项目，可以不传
func Watch(name, root, goWay, mainFile string, deamon bool, depends ...string) error {
    prj, err := New(name, root, goWay, mainFile, deamon, depends...)
    if err != nil {
        return err
    }
    if err = prj.CreateMakeFile(); err != nil {
        log.Println("create make file error:", err)
        return err
    }
    defer prj.Watch()
    if goWay == "run" {
        return prj.Run()
    }
    if err = prj.Compile(); err != nil {
        return err
    }
    if err = prj.Start(); err != nil {
        return err
    }
    log.Println("[INFO] 项目", name, "启动完成")
    return nil
}

type Project struct {
    name            string   // 项目名称
    Root            string   // 项目的根路径
    binAbsolutePath string   // 执行文件路径（绝对路径）
    execArgs        []string // 程序执行的参数
    srcAbsolutePath string   // 源程序文件路径（绝对路径）
    errAbsolutePath string   // 编译语法错误存放位置

    GoWay   string // 项目编译方式:run、build还是install
    deamon  bool   // 程序是否一直运行（比如Web服务）
    Options string // 编译选项

    // 对于run、build而言，是main包中的main函数所在路径（包括文件名）
    // 对于install而言，是main包所在目录（可能多级），如果没配置，则等于name
    MainFile string
    Depends  []string // 依赖其他项目（一般只是库）

    process *os.Process
}

// New 要求被监听项目必须有src目录（按Go习惯建目录）
func New(name, root, goWay, mainFile string, deamon bool, depends ...string) (*Project, error) {
    if !files.IsDir(root) {
        return nil, PrjRootErr
    }
    if filepath.IsAbs(mainFile) {
        return nil, errors.New("main配置项必须是相对于项目src的相对路径！")
    }
    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }
    binAbsolutePath := filepath.Join(root, "bin")

    options := ""
    switch goWay {
    case "run":
        mainFile = filepath.Join("src", mainFile)
    case "build":
        mainFile = filepath.Join("src", mainFile)
        if !files.Exist(binAbsolutePath) {
            if err = os.Mkdir(binAbsolutePath, 0777); err != nil {
                return nil, err
            }
        }
        output := filepath.Join("bin", name+binanryFileSuffix)
        options = "-o " + output
    case "install":
        fallthrough
    default:
        if mainFile == "" {
            mainFile = name
        } else {
            mainFile = filepath.Dir(mainFile)
        }
    }
    return &Project{
        name:            name,
        Root:            root,
        binAbsolutePath: binAbsolutePath,
        srcAbsolutePath: filepath.Join(root, "src"),
        errAbsolutePath: filepath.Join(root, "_log_"),
        GoWay:           goWay,
        deamon:          deamon,
        MainFile:        mainFile,
        Options:         options,
        Depends:         depends,
    }, nil
}

// Watch 监听该项目，源码有改动会重新编译运行
func (this *Project) Watch() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    eventNum := make(chan int)
    go func() {
        for {
            i := 0
        GetEvent:
            for {
                select {
                case <-watcher.Event:
                    i++
                // 修改可能会有多次modify事件
                case <-time.After(500e6):
                    break GetEvent
                }
            }
            if i > 0 {
                eventNum <- i
            }
        }
    }()

    go func() {
        for {
            var err error
            select {
            case <-eventNum:
                if this.GoWay == "run" {
                    if err = this.Run(); err != nil {
                        log.Println("run error:", err)
                    }
                    break
                }
                if err = this.Compile(); err != nil {
                    log.Println("complie error: ", err)
                    break
                }
                if err = this.Restart(); err != nil {
                    log.Println("restart error:", err)
                }
            }
            if this.deamon && err == nil {
                log.Println("重启完成！")
            }
        }
    }()

    addWatch(watcher, this.srcAbsolutePath)
    return nil
}

// addWatch 使用fsnotify，监听src目录以及子目录
func addWatch(watcher *fsnotify.Watcher, dir string) {
    watcher.Watch(dir)
    for _, filename := range files.ScanDir(dir) {
        childDir := filepath.Join(dir, filename)
        if files.IsDir(childDir) {
            addWatch(watcher, childDir)
        }
    }
}

// SetDepends 设置依赖的项目，被依赖的项目一般是tools
func (this *Project) SetDepends(depends ...string) {
    for _, depend := range depends {
        this.Depends = append(this.Depends, depend)
    }
}

// ChangetoRoot 切换到当前Project的根目录
func (this *Project) ChangeToRoot() error {
    if err := os.Chdir(this.Root); err != nil {
        log.Println(err)
        return err
    }
    return nil
}

// CreateMakeFile 创建make文件（在当前工程根目录），这里的make文件和makefile不一样
// 这里的make文件只是方便编译当前工程而不依赖于GOPATH
func (this *Project) CreateMakeFile() error {
    // 获得当前目录
    path, err := os.Getwd()
    if err != nil {
        return err
    }
    this.ChangeToRoot()
    file, err := os.Create(filepath.Join(this.Root, installFileName))
    if err != nil {
        os.Chdir(path)
        return err
    }
    os.Chdir(path)
    defer file.Close()
    tpl := template.Must(template.ParseFiles(makeTplFile))
    tpl.Execute(file, this)
    return nil
}

// Run 当GoWay==run时，直接通过该方法，而不需要先Compile然后Start
func (this *Project) Run() error {
    path, err := os.Getwd()
    if err != nil {
        return err
    }
    this.ChangeToRoot()
    defer os.Chdir(path)
    os.Chmod(installFileName, 0755)
    cmd := exec.Command(installCmd)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err = cmd.Start(); err != nil {
        return err
    }
    // TODO:据说time.Sleep会内存泄露
    select {
    case <-time.After(300e6):
    }
    output := strings.TrimSpace(stdout.String())
    errOutput := strings.TrimSpace(stderr.String())
    errFile := filepath.Join(this.errAbsolutePath, "error.html")
    if this.deamon {
        if output == "" {
            // 删除可能的错误文件夹和文件
            if files.Exist(errFile) {
                os.RemoveAll(this.errAbsolutePath)
            }
            this.process = cmd.Process
            return nil
        }
    } else {
        log.Println("=====================")
        log.Println("[INFO] 项目", this.name, "的运行结果:")
        for _, val := range strings.Split(errOutput, "\n") {
            log.Println(val)
        }
        log.Println("=====================")
        return nil
    }

    if strings.Contains(errOutput, "listen tcp") {
        if err = this.process.Kill(); err == nil {
            this.Run()
        }
        return nil
    }

    // 往项目中写入错误信息
    if !files.Exist(this.errAbsolutePath) {
        if err = os.Mkdir(this.errAbsolutePath, 0777); err != nil {
            log.Println("can't create errAbsolutePath: ", err)
        }
    }
    file, err := os.Create(errFile)
    if err != nil {
        return err
    }
    defer file.Close()
    tpl.Execute(file, errOutput)
    return PrjSyntaxError
}

// Compile 编译当前Project。
func (this *Project) Compile() error {
    path, err := os.Getwd()
    if err != nil {
        return err
    }
    this.ChangeToRoot()
    defer os.Chdir(path)
    // 删除bin中的文件
    if this.GoWay == "build" {
        binFile := this.getExeFilePath()
        if files.Exist(binFile) {
            os.Remove(binFile)
        }
    }
    os.Chmod(installFileName, 0755)
    cmd := exec.Command(installCmd)
    var stdout bytes.Buffer
    cmd.Stdout = &stdout
    if err = cmd.Run(); err != nil {
        return err
    }
    output := strings.TrimSpace(stdout.String())
    errFile := filepath.Join(this.errAbsolutePath, "error.html")
    if successFlag == output {
        // 删除可能的错误文件夹和文件
        if files.Exist(errFile) {
            os.RemoveAll(this.errAbsolutePath)
        }
        return nil
    }

    // 往项目中写入错误信息
    if !files.Exist(this.errAbsolutePath) {
        if err = os.Mkdir(this.errAbsolutePath, 0777); err != nil {
            log.Println("can't create errAbsolutePath: ", err)
        }
    }
    file, err := os.Create(errFile)
    if err != nil {
        return err
    }
    defer file.Close()
    output = strings.Replace(output, "finished", "", -1)
    tpl.Execute(file, output)
    return PrjSyntaxError
}

// Start 启动该Project
func (this *Project) Start() error {
    path, err := os.Getwd()
    if err != nil {
        return err
    }
    this.ChangeToRoot()
    defer os.Chdir(path)
    cmd := exec.Command(this.getExeFilePath(), this.execArgs...)
    return cmd.Start()
}

// 重新启动该Project
func (this *Project) Restart() error {
    if err := this.Stop(); err != nil {
        log.Println("stop project error!", err)
        return err
    }
    return this.Start()
}

// getExeFilePath 获得可执行文件路径（项目）
func (this *Project) getExeFilePath() string {
    return filepath.Join(this.binAbsolutePath, this.MainFile+binanryFileSuffix)
}
