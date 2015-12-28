// Build constraints copied from go's src/os/dir_unix.go
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import (
	"os/user"
	"path/filepath"
)

var defaultConfigPaths = []string{
	// This will be prepended by $HOME/.config/gobuster.conf
	"/etc/gobuster.conf",
}

func init() {
	if usr, err := user.Current(); err != nil {
		path := filepath.Join(usr.HomeDir, ".config", "gobuster.conf")
		defaultConfigPaths = append([]string{path}, defaultConfigPaths...)
	}
}
