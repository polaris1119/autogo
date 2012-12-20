autogo
======

Go语言是静态语言，修改源代码总是需要编译、运行，如果用Go做Web开发，修改一点就要编译、运行，然后才能看结果，很痛苦。
autogo就是为了让Go开发更方便。在开发阶段，做到修改之后，立马看到修改的效果，如果编译出错，能够及时显示错误信息！

使用说明
======

1、下载
将源代码git clone到任意一个位置

2、修改config/project.json文件
  该文件的作用是配置需要被autogo管理的项目，一个项目是一个json对象（{}），包括name、root和depends，
  其中depends是可选的，name是项目最终生成可执行文件的文件名（也就是main包所在的目录）；root是项目的根目录。

3、执行make(linux)/make.bat(windows)，编译autogo

4、运行autogo：bin/autogo
  注意，运行autogo时，当前目录要切换到autogo所在目录
  
注：为了方便编译出错时看到错误详细信息，当有错误时autogo会在项目中新建一个文件，将错误信息写入其中。
因此建议测阶段，在被监控的项目中加入如下一段代码（在所有访问的入口处）：
    
    errFile := "_log_/error.html"
    _, err := os.Stat(errFile)
    if err == nil || os.IsExist(err) {
        content, _ := ioutil.ReadFile(errFile)
        fmt.Fprintln(rw, string(content))
        return
    }
这样，当程序编译出错时，刷新页面会看到类似如下的错误提示：

    ~~o(>_<)o ~~主人，编译出错了哦！
    
    错误详细信息：
    
    # test src\test\main.go:5: imported and not used: "io"

例子程序
======

1、在任意目录新建一个test工程。目录结构如下：
    
    test
    └───src
        └───test
              └───main.go
2、main.go的代码如下：
    
    import (
      "fmt"
      "io/ioutil"
      "log"
      "net/http"
      "os"
      "runtime"
    )
    func init() {
        runtime.GOMAXPROCS(runtime.NumCPU())
    }
    
    func main() {
        http.HandleFunc("/", mainHandle)
    
        log.Fatal(http.ListenAndServe(":8080", nil))
    }
    
    func mainHandle(rw http.ResponseWriter, req *http.Request) {
        // 当编译出错时提示错误信息；开发阶段使用
        errFile := "_log_/error.html"
        _, err := os.Stat(errFile)
        if err == nil || os.IsExist(err) {
            content, _ := ioutil.ReadFile(errFile)
            fmt.Fprintln(rw, string(content))
            return
        }
        fmt.Fprintln(rw, "Hello, World!")
        // 这里可以统一路由转发
    }

3、在autogo的config/project.json中将该项目加进去
    
    [
      {
          "name": "test",
          "root": "../test",
          "depends": []
      }
    ]
    root可以是相对路径或决定路径、

4、启动autogo（如果autogo没编译，先通过make编译）。注意，启动autogo应该cd到autogo所在根目录执行bin/autogo启动。

5、在浏览器中访问：http://localhost:8080，就可以看到Hello World！了。
  改动test中的main.go，故意出错，然后刷新页面看到效果了有木有！

版本更新历史
=====

2012-12-20  autogo 2.0发布
```
1、优化编译、运行过程（只会执行一次）
2、支持多种goway方式：go run、build、install，这样对于测试项目也支持了
3、修复了 被监控项目如果有问题 autogo启动不了的情况
4、调整了代码结构
```

2012-12-18  autogo 1.0发布

使用的第三方库
======

为了方便，autogo中直接包含了第三方库，不需要另外下载。

1、[fsnotify](https://github.com/howeyc/fsnotify)，File system notifications

2、[simplejson](https://github.com/bitly/go-simplejson)，解析JSON，我做了一些改动

感谢
=====

johntech

[ohlinux](https://github.com/ohlinux)

LICENCE
======

The MIT [License](https://github.com/polaris1119/autogo/master/LICENSE)
