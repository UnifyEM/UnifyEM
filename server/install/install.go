/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package install

import (
	"fmt"
	"io"
	"os"

	"github.com/UnifyEM/UnifyEM/server/global"
)

type Install struct {
	conf *global.ServerConfig
}

func New(conf *global.ServerConfig) *Install {
	return &Install{conf: conf}
}

// Check displays the current configuration
func (i *Install) Check() {
	c, err := i.conf.SC.Dump()
	if err != nil {
		fmt.Printf("Error dumping configuration: %v\n", err)
		return
	}
	fmt.Printf("Server Configuration:\n\n%s\n", c)
}

func (i *Install) Install() error {
	var err error

	// Call the private function for os specific install
	err = i.installService()
	if err != nil {
		return err
	}

	// Save the config
	return i.conf.Checkpoint()
}

func (i *Install) Uninstall() error {
	// Call the private function for os specific uninstall
	return i.uninstallService(true)
}

func (i *Install) Upgrade() error {
	// Call the private function for os specific upgrade
	return i.upgradeService()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	return nil
}
