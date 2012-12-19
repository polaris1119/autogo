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
// depends：是依赖的其他GOPATH路径下的项目，可以不传
func Watch(name, root string, depends ...string) error {
    prj, err := New(name, root, depends...)
    if err != nil {
        return err
    }
    if err = prj.CreateMakeFile(); err != nil {
        log.Println("create make file error:", err.Error())
        return err
    }
    if !files.Exist(prj.errAbsolutePath) {
        os.Mkdir(prj.errAbsolutePath, 0777)
    }
    if err = prj.Compile(); err != nil {
        return err
    }
    if err = prj.Start(); err != nil {
        return err
    }
    return prj.Watch()
}

type Project struct {
    // 项目名称，要求包含main包的文件夹是这个名字
    Name string
    // 项目的根路径
    Root string
    // 执行文件路径（绝对路径）
    binAbsolutePath string
    // 程序执行的参数
    execArgs []string
    // 源程序文件路径（绝对路径）
    srcAbsolutePath string
    // 编译语法错误存放位置
    errAbsolutePath string

    // 依赖其他项目（一般只是库）
    Depends []string
}

// New 要求被监听项目必须有src目录（按Go习惯建目录）
func New(name, root string, depends ...string) (*Project, error) {
    if !files.IsDir(root) {
        return nil, PrjRootErr
    }
    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }
    return &Project{
        Name:            name,
        Root:            root,
        binAbsolutePath: filepath.Join(root, "bin"),
        srcAbsolutePath: filepath.Join(root, "src"),
        errAbsolutePath: filepath.Join(root, "_log_"),
        Depends:         depends,
    }, nil
}

// Watch 监听该项目，源码有改动会重新编译运行
func (this *Project) Watch() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    go func() {
        for {
            select {
            case event := <-watcher.Event:
                log.Println(event)
                this.Compile()
                if err = this.Restart(); err != nil {
                    log.Println("restart error:", err)
                }
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
    file, err := os.Create(filepath.Join(this.Root, makeFileName))
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

// Compile 编译当前Project
func (this *Project) Compile() error {
    path, err := os.Getwd()
    if err != nil {
        return err
    }
    this.ChangeToRoot()
    defer os.Chdir(path)
    os.Chmod(makeFileName, 0755)
    cmd := exec.Command(os.Getenv("SHELL"), "-c", "cd "+this.Root+" && ./"+makeFileName)
    var out bytes.Buffer
    cmd.Stdout = &out
    if err := cmd.Run(); err != nil {
        return err
    }
    log.Println(out.String())
    output := strings.TrimSpace(out.String())
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
        os.Mkdir(this.errAbsolutePath, 0777)
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
    return filepath.Join(this.binAbsolutePath, this.Name+binanryFileSuffix)
}
