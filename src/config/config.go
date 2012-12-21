package config

import (
    "fsnotify"
    "log"
    "project"
    "simplejson"
    "time"
)

// Watch 监控配置文件
func Watch(configFile string) error {
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
                case <-time.After(200e6):
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
            select {
            case <-eventNum:
                log.Println("[INFO] ReloadConfig...")
                Load(configFile)
            }
        }
    }()

    return watcher.Watch(configFile)
}

// Load加载解析配置文件
func Load(configFile string) error {
    allConfig, err := simplejson.ParseFile(configFile)
    if err != nil {
        log.Println("[ERROR] 配置文件格式错误", err)
        return err
    }
    middleJs, err := allConfig.Array()
    if err != nil {
        log.Println("[ERROR] 配置文件格式错误", err)
        return err
    }
    for i, length := 0, len(middleJs); i < length; i++ {
        oneProject := allConfig.GetIndex(i)
        name := oneProject.Get("name").MustString()
        root := oneProject.Get("root").MustString()
        goWay := oneProject.Get("go_way").MustString()
        deamon := oneProject.Get("deamon").MustBool(true)
        mainFile := oneProject.Get("main").MustString()
        depends := oneProject.GetStringSlice("depends")
        err = project.Watch(name, root, goWay, mainFile, deamon, depends...)
        if err != nil {
            log.Println("[ERROR] 监控Project：", name, " 出错：", err)
        }
    }
    return err
}