// Copyright 2012 polaris(studygolang.com). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
    "config"
    "flag"
    "runtime"
)

var configFile string

func init() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    flag.StringVar(&configFile, "f", "config/projects.json", "配置文件：需要监听哪些工程")
    flag.Parse()
}

func main() {
    config.Load(configFile)
    config.Watch(configFile)
    select {}
}
