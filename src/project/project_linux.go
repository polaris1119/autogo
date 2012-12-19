// Copyright 2012 polaris(studygolang.com). All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package project

import (
    "os/exec"
)

var (
    makeTplFile       = "templates/make_linux.tpl"
    makeFileName      = "run"
    binanryFileSuffix = ""
)

// Stop 停止该Project
func (this *Project) Stop() error {
    cmd := exec.Command("killall", this.Name)
    if err := cmd.Run(); err != nil {
        return err
    }
    return nil
}
