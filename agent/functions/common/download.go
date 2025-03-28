//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package common

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

func DownloadExecute(logger interfaces.Logger, comms *communications.Communications, url string, args []string) error {

	// Download the file
	tmpFile, err := comms.Download(url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}

	logger.Infof(8101, "Downloaded %s to %s", url, tmpFile)

	// On Windows, we need to add the .exe extension to the file
	if runtime.GOOS == "windows" {
		// if the file does not end in .exe, add it
		if len(tmpFile) < 4 || tmpFile[len(tmpFile)-4:] != ".exe" {
			// rename the file to add the .exe extension
			newName := tmpFile + ".exe"
			err = os.Rename(tmpFile, newName)
			if err != nil {
				_ = os.Remove(tmpFile)
				return fmt.Errorf("error renaming %s to %s: %w", tmpFile, newName, err)
			}
			tmpFile = newName
		}
	}

	// Make the file executable
	err = os.Chmod(tmpFile, 0755)
	if err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("error making %s executable: %w", tmpFile, err)
	}

	logger.Infof(8102, "executing %s with argument(s) %v", tmpFile, args)

	// Execute the file with the supplied arguments
	cmd := exec.Command(tmpFile, args...)

	// Set platform-specific process attributes so that the child process can continue after
	// the parent process (which one) exists
	setProcessAttributes(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command and do not wait for it to complete
	err = cmd.Start()
	if err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("error executing %s: %w", tmpFile, err)
	}

	// Do not wait for the command to complete
	logger.Infof(8103, "Started %s with argument(s) %v", tmpFile, args)
	return nil
}
