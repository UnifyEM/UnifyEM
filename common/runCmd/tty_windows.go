//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Windows-specific constants for CreateProcessWithLogonW
const (
	LOGON_WITH_PROFILE         = 0x00000001
	CREATE_UNICODE_ENVIRONMENT = 0x00000400
)

// TTY runs a command in Windows, handling interactive prompts where possible.
// If AsUser is specified, it will run the command as that user using Windows authentication.
// Returns all output from the command and any error that occurred.
func osTTY(def Interactive) (string, error) {
	if def.AsUser != nil {
		// Use Windows-specific user impersonation
		return runAsUserWindows(def)
	}

	// For normal execution without user switch
	// Windows doesn't have PTY support, so we use pipes for interaction
	return runInteractiveWindows(def)
}

// runAsUserWindows runs a command as a different user on Windows
func runAsUserWindows(def Interactive) (string, error) {
	// Method 1: Use runas command (simpler but requires interaction)
	if len(def.Actions) == 0 {
		// For non-interactive commands, use CreateProcessWithLogonW via PowerShell
		return runWithPowerShellCredentials(def)
	}

	// Method 2: Use runas with interaction handling
	return runWithRunas(def)
}

// runWithPowerShellCredentials uses PowerShell's Start-Process with credentials
func runWithPowerShellCredentials(def Interactive) (string, error) {
	// Create a PowerShell script that runs the command with credentials
	psScript := fmt.Sprintf(`
$secpasswd = ConvertTo-SecureString '%s' -AsPlainText -Force
$creds = New-Object System.Management.Automation.PSCredential ('%s', $secpasswd)
$pinfo = New-Object System.Diagnostics.ProcessStartInfo
$pinfo.FileName = '%s'
$pinfo.Arguments = '%s'
$pinfo.RedirectStandardOutput = $true
$pinfo.RedirectStandardError = $true
$pinfo.UseShellExecute = $false
$pinfo.UserName = '%s'
$pinfo.Password = $secpasswd
$pinfo.Domain = $env:COMPUTERNAME
$p = New-Object System.Diagnostics.Process
$p.StartInfo = $pinfo
$p.Start() | Out-Null
$stdout = $p.StandardOutput.ReadToEnd()
$stderr = $p.StandardError.ReadToEnd()
$p.WaitForExit()
Write-Output $stdout
if ($stderr) { Write-Error $stderr }
exit $p.ExitCode
`, def.AsUser.Password, def.AsUser.Username,
		def.Command[0],
		strings.Join(def.Command[1:], " "),
		def.AsUser.Username)

	// Run the PowerShell script
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// runWithRunas uses the runas command for interactive scenarios
func runWithRunas(def Interactive) (string, error) {
	// Build command for runas
	cmdStr := strings.Join(def.Command, " ")
	runasCmd := exec.Command("cmd", "/C",
		fmt.Sprintf(`echo %s | runas /user:%s "%s"`,
			def.AsUser.Password, def.AsUser.Username, cmdStr))

	// Set up pipes for stdin/stdout/stderr
	stdin, err := runasCmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	var outputBuf bytes.Buffer
	runasCmd.Stdout = &outputBuf
	runasCmd.Stderr = &outputBuf

	// Start the command
	if err := runasCmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start runas: %w", err)
	}

	// Handle any additional interactions
	for _, action := range def.Actions {
		if action.WaitFor != "" {
			// Wait for expected output
			deadline := time.Now().Add(30 * time.Second)
			for time.Now().Before(deadline) {
				if strings.Contains(outputBuf.String(), action.WaitFor) {
					time.Sleep(250 * time.Millisecond)
					stdin.Write([]byte(action.Send + "\r\n"))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// Wait for completion
	err = runasCmd.Wait()
	return outputBuf.String(), err
}

// runInteractiveWindows handles interactive commands without user switching
func runInteractiveWindows(def Interactive) (string, error) {
	cmd := exec.Command(def.Command[0], def.Command[1:]...)

	// Set up pipes for stdin/stdout/stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Handle interactions
	for _, action := range def.Actions {
		if action.WaitFor != "" {
			// Wait for expected output
			deadline := time.Now().Add(30 * time.Second)
			for time.Now().Before(deadline) {
				if strings.Contains(outputBuf.String(), action.WaitFor) {
					time.Sleep(250 * time.Millisecond)
					stdin.Write([]byte(action.Send + "\r\n"))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// Wait for completion
	err = cmd.Wait()
	return outputBuf.String(), err
}

// Alternative: Use Windows API directly (more complex but more reliable)
func runAsUserWindowsAPI(username, password, domain string, command []string) (string, error) {
	// This would use syscall to call CreateProcessWithLogonW directly
	// This is more complex but gives better control

	if domain == "" {
		domain = "." // Local computer
	}

	_ = syscall.StringToUTF16Ptr(strings.Join(command, " "))

	var si syscall.StartupInfo
	var _ syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))

	// Note: This is pseudo-code - you'd need to import the correct Windows API
	// kernel32 := syscall.NewLazyDLL("advapi32.dll")
	// createProcessWithLogonW := kernel32.NewProc("CreateProcessWithLogonW")

	// The actual API call would go here
	// This is quite complex to implement correctly

	return "", fmt.Errorf("direct API not implemented - use PowerShell method")
}
