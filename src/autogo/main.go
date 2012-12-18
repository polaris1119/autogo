// Copyright 2012 polaris(studygolang.com). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
    "flag"
    "fsnotify"
    "io/ioutil"
    "log"
    "project"
    "runtime"
    "simplejson"
)

var prj *project.Project
var err error

var configFile string

func init() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    flag.StringVar(&configFile, "f", "config/projects.json", "配置文件：需要监听哪些工程")
    flag.Parse()
}

func main() {
    if err := LoadConfig(); err != nil {
        log.Fatal(err)
    }
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    go func() {
        for {
            select {
            case <-watcher.Event:
                log.Println("ReloadConfig...")
                LoadConfig()
            }
        }
    }()
    watcher.Watch(configFile)
    if err != nil {
        log.Fatal(err)
    }
    select {}
}

func LoadConfig() error {
    content, err := ioutil.ReadFile(configFile)
    if err != nil {
        log.Fatal("读配置文件错误")
    }
    allConfig, err := simplejson.NewJson(content)
    if err != nil {
        log.Fatal("配置文件格式错误")
    }
    middleJs, err := allConfig.Array()
    if err != nil {
        log.Fatal("配置文件格式错误")
    }
    for i, length := 0, len(middleJs); i < length; i++ {
        oneProject := allConfig.GetIndex(i)
        name := oneProject.Get("name").MustString()
        root := oneProject.Get("root").MustString()
        depends := oneProject.GetStringSlice("depends")
        err = project.Watch(name, root, depends...)
    }
    return err
}
