//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package common

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/execute"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/hasher"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

func Download(logger interfaces.Logger, comms *communications.Communications, url string, hash string) (string, error) {

	// Check for the (usually) required hash
	if hash == "" {
		if global.DisableHash {
			logger.Warningf(8101, "No hash provided, but continuing because hash verification is disabled")
		} else {
			return "", fmt.Errorf("empty hash string received, refusing to download %s", url)
		}
	}

	// Download the file
	tmpFile, err := comms.Download(url)
	if err != nil {
		return "", fmt.Errorf("error downloading %s: %w", url, err)
	}

	logger.Infof(8102, "Downloaded %s to %s", url, tmpFile)

	// Verify the hash
	h := hasher.New()
	if !h.SHA256File(tmpFile).Compare(hash) {
		if global.DisableHash {
			logger.Warningf(8103, "Hash verification failed, but continuing because hash verification is disabled")
		} else {
			_ = os.Remove(tmpFile)
			return "", fmt.Errorf("hash verification failed, deleted %s", tmpFile)
		}
	}

	logger.Infof(8104, "Hash verification succeeded for %s", tmpFile)
	return tmpFile, nil
}

func DownloadExecute(logger interfaces.Logger, comms *communications.Communications, url string, args []string, hash string) error {

	// Download the file
	tmpFile, err := Download(logger, comms, url, hash)
	if err != nil {
		return err
	}

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

	logger.Infof(8105, "executing %s with argument(s) %v", tmpFile, args)
	return execute.Execute(logger, tmpFile, args)
}

// FileToMap reads the specified JSON file and returns a map[string]string of the contents
func FileToMap(file string) (map[string]string, error) {
	// Open the file
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", file, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Decode the JSON file into a map
	var data map[string]string
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("error deserializing %s: %w", file, err)
	}
	return data, nil
}
