//
//  Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package status

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// osName returns the OS name
func (h *Handler) osName() string {
	return "Linux"
}

// osVersion returns the Linux distribution and version
func (h *Handler) osVersion() string {
	// Try /etc/os-release (standard on all modern distros)
	f, err := os.Open("/etc/os-release")
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(line[13:], "\"")
			}
		}
	}
	// Fallback to /etc/redhat-release or /etc/centos-release (older RHEL/CentOS)
	for _, relFile := range []string{"/etc/redhat-release", "/etc/centos-release"} {
		if f, err := os.Open(relFile); err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			if scanner.Scan() {
				return strings.TrimSpace(scanner.Text())
			}
		}
	}
	// Fallback to lsb_release -d (may not be installed by default)
	out, err := exec.Command("lsb_release", "-d").Output()
	if err == nil {
		parts := strings.SplitN(string(out), ":", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	return "unknown"
}

// firewall returns "yes" if a firewall is enabled, "no" if not, "unknown" otherwise
func (h *Handler) firewall() string {
	// Check for ufw
	out, err := exec.Command("ufw", "status").Output()
	if err == nil {
		if bytes.Contains(out, []byte("Status: active")) {
			return "yes"
		}
		if bytes.Contains(out, []byte("Status: inactive")) {
			return "no"
		}
	}
	// Check for firewalld
	out, err = exec.Command("systemctl", "is-active", "firewalld").Output()
	if err == nil {
		if strings.TrimSpace(string(out)) == "active" {
			return "yes"
		}
		if strings.TrimSpace(string(out)) == "inactive" {
			return "no"
		}
	}
	// Check for iptables rules
	out, err = exec.Command("iptables", "-L").Output()
	if err == nil && len(out) > 0 {
		// If there are any rules other than default ACCEPT, assume firewall is active
		if !bytes.Contains(out, []byte("Chain INPUT (policy ACCEPT)")) {
			return "yes"
		}
	}
	return "unknown"
}

// antivirus returns "yes" if a known AV process is running, "no" if not, "unknown" otherwise
func (h *Handler) antivirus() string {
	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return "unknown"
	}
	for _, proc := range antivirusProcesses {
		if bytes.Contains(out, []byte(proc)) {
			return "yes"
		}
	}
	return "no"
}

// autoUpdates returns "yes" if automatic updates are enabled, "no" if not, "unknown" otherwise
func (h *Handler) autoUpdates() string {
	// Check for unattended-upgrades (Debian/Ubuntu)
	f, err := os.Open("/etc/apt/apt.conf.d/20auto-upgrades")
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "APT::Periodic::Unattended-Upgrade") {
				if strings.Contains(line, "1") {
					return "yes"
				}
				return "no"
			}
		}
	}
	// Check for dnf-automatic (Fedora/RHEL)
	out, err := exec.Command("systemctl", "is-enabled", "dnf-automatic.timer").Output()
	if err == nil {
		if strings.Contains(string(out), "enabled") {
			return "yes"
		}
		return "no"
	}
	return "unknown"
}

// fde returns "yes" if full disk encryption is enabled, "no" if not, "unknown" otherwise
func (h *Handler) fde() string {
	// Check if root is on a LUKS device
	out, err := exec.Command("lsblk", "-o", "NAME,TYPE,MOUNTPOINT").Output()
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "/") && strings.Contains(line, "crypt") {
			return "yes"
		}
	}
	// Check for LUKS header
	out, err = exec.Command("cryptsetup", "status", "root").Output()
	if err == nil {
		if strings.Contains(string(out), "LUKS header") {
			return "yes"
		}
	}
	// Check for dm-crypt
	out, err = exec.Command("ls", "/dev/mapper").Output()
	if err == nil {
		if bytes.Contains(out, []byte("cryptroot")) || bytes.Contains(out, []byte("cryptswap")) {
			return "yes"
		}
	}
	// Check for eCryptfs
	out, err = exec.Command("mount").Output()
	if err == nil {
		if bytes.Contains(out, []byte("ecryptfs")) {
			return "yes"
		}
	}
	return "no"
}

// password returns "yes" if the current user has a password set, "no" if not, "unknown" otherwise
func (h *Handler) password() string {
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		// Try LOGNAME as a fallback
		currentUser = os.Getenv("LOGNAME")
	}
	if currentUser == "" {
		return "unknown"
	}

	f, err := os.Open("/etc/shadow")
	if err != nil {
		return "unknown"
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, currentUser+":") {
			fields := strings.Split(line, ":")
			if len(fields) > 1 {
				hash := fields[1]
				if hash == "!" || hash == "*" || hash == "" {
					return "no"
				}
				return "yes"
			}
		}
	}
	return "unknown"
}

// getDisplayEnv tries to find a DISPLAY environment variable for a running X11/Wayland session.
// Returns the DISPLAY value and true if found, otherwise "" and false.
func (h *Handler) getDisplayEnv() (string, bool) {
	// Check for X11 sockets
	if files, err := os.ReadDir("/tmp/.X11-unix"); err == nil && len(files) > 0 {
		// Try to find a DISPLAY from a running Xorg/X process
		out, err := exec.Command("ps", "axo", "pid,comm").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) == 2 && (fields[1] == "Xorg" || fields[1] == "X") {
					pid := fields[0]
					environPath := "/proc/" + pid + "/environ"
					if envBytes, err := os.ReadFile(environPath); err == nil {
						envVars := strings.Split(string(envBytes), "\x00")
						for _, env := range envVars {
							if strings.HasPrefix(env, "DISPLAY=") {
								return strings.TrimPrefix(env, "DISPLAY="), true
							}
						}
					}
				}
			}
		}
		// Fallback: assume :0 if X11 socket exists
		return ":0", true
	}
	// Check for Wayland socket for any user
	runUserDirs, _ := filepath.Glob("/run/user/*")
	for _, dir := range runUserDirs {
		waylandSock := filepath.Join(dir, "wayland-0")
		if _, err := os.Stat(waylandSock); err == nil {
			// Try to find WAYLAND_DISPLAY from a compositor process
			out, err := exec.Command("ps", "axo", "pid,comm").Output()
			if err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					fields := strings.Fields(line)
					if len(fields) == 2 && (strings.Contains(fields[1], "wayland") || strings.Contains(fields[1], "gnome-session")) {
						pid := fields[0]
						environPath := "/proc/" + pid + "/environ"
						if envBytes, err := ioutil.ReadFile(environPath); err == nil {
							envVars := strings.Split(string(envBytes), "\x00")
							for _, env := range envVars {
								if strings.HasPrefix(env, "WAYLAND_DISPLAY=") {
									return strings.TrimPrefix(env, "WAYLAND_DISPLAY="), true
								}
							}
						}
					}
				}
			}
			// Fallback: assume wayland-0
			return "wayland-0", true
		}
	}
	return "", false
}

// screenLock returns "yes" if the user's screen will automatically lock after inactivity, "no" if not, "unknown" otherwise
func (h *Handler) screenLock() (string, error) {
	display, found := h.getDisplayEnv()
	if !found {
		return "n/a", nil
	}
	// Set DISPLAY for child commands
	os.Setenv("DISPLAY", display)
	// Check GNOME settings: lock-enabled and idle-delay
	lockOut, err1 := exec.Command("gsettings", "get", "org.gnome.desktop.screensaver", "lock-enabled").Output()
	idleOut, err2 := exec.Command("gsettings", "get", "org.gnome.desktop.session", "idle-delay").Output()
	if err1 == nil && err2 == nil {
		lockVal := strings.TrimSpace(string(lockOut))
		idleVal := strings.Trim(strings.TrimSpace(string(idleOut)), "'")
		if lockVal == "true" {
			// idle-delay is in seconds, must be > 0
			if idleSec, err := strconv.Atoi(idleVal); err == nil && idleSec > 0 {
				return "yes", nil
			}
			return "no", nil
		}
		return "no", nil
	}

	// Try xdg-screensaver (generic X11) as a fallback
	out, err := exec.Command("xdg-screensaver", "status").Output()
	if err == nil {
		if strings.Contains(string(out), "enabled") {
			return "yes", nil
		}
		if strings.Contains(string(out), "disabled") {
			return "no", nil
		}
	}
	return "unknown", nil
}

func (h *Handler) screenLockDelay() string {
	display, found := h.getDisplayEnv()
	if !found {
		return "n/a"
	}
	// Set DISPLAY for child commands
	os.Setenv("DISPLAY", display)
	// Try gsettings (GNOME, Ubuntu, Debian, CentOS default)
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.session", "idle-delay").Output()
	if err == nil {
		val := strings.TrimSpace(string(out))
		val = strings.Trim(val, "'")
		return val
	}

	// Try xfconf-query (XFCE)
	out, err = exec.Command("xfconf-query", "-c", "xfce4-session", "-p", "/general/LockCommand").Output()
	if err == nil {
		val := strings.TrimSpace(string(out))
		if val != "" {
			return val
		}
	}

	// Try KDE config (~/.config/kscreenlockerrc)
	home := os.Getenv("HOME")
	if home != "" {
		kdeConf := home + "/.config/kscreenlockerrc"
		f, err := os.Open(kdeConf)
		if err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "Timeout=") {
					val := strings.TrimPrefix(line, "Timeout=")
					return val
				}
			}
		}
	}

	return "unknown"
}

// lastUser returns the last logged-in user
func (h *Handler) lastUser() string {
	out, err := exec.Command("last", "-w").Output()
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] != "reboot" && fields[0] != "wtmp" {
			return fields[0]
		}
	}
	return "unknown"
}

// bootTime returns the system boot time as an ISO8601 string
func (h *Handler) bootTime() string {
	// Try /proc/stat for btime
	f, err := os.Open("/proc/stat")
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "btime") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					epoch, err := strconv.ParseInt(parts[1], 10, 64)
					if err == nil {
						t := time.Unix(epoch, 0)
						return t.Format("2006-01-02T15:04:05-07:00")
					}
				}
			}
		}
	}
	// Fallback to uptime
	out, err := exec.Command("uptime", "-s").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return "unknown"
}
