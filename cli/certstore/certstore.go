/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package certstore

import (
	"bufio"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const certFile = ".uemcert"

// Fingerprint computes the SHA-256 fingerprint of a DER-encoded certificate.
func Fingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return fmt.Sprintf("%X", hash)
}

// FormatCertDetails returns a human-readable summary of a certificate.
func FormatCertDetails(cert *x509.Certificate) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  Subject:     %s\n", cert.Subject))
	b.WriteString(fmt.Sprintf("  Issuer:      %s\n", cert.Issuer))
	b.WriteString(fmt.Sprintf("  Fingerprint: %s\n", Fingerprint(cert)))
	b.WriteString(fmt.Sprintf("  Not Before:  %s\n", cert.NotBefore.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("  Not After:   %s\n", cert.NotAfter.Format(time.RFC3339)))
	if len(cert.DNSNames) > 0 {
		b.WriteString(fmt.Sprintf("  DNS Names:   %s\n", strings.Join(cert.DNSNames, ", ")))
	}
	return b.String()
}

// certFilePath returns the full path to ~/.uemcert.
func certFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(homeDir, certFile), nil
}

// IsTrusted checks whether the given host and fingerprint are stored in ~/.uemcert.
func IsTrusted(host, fingerprint string) (bool, error) {
	path, err := certFilePath()
	if err != nil {
		return false, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("unable to open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == host && parts[1] == fingerprint {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// Store upserts a host and fingerprint entry in ~/.uemcert.
// If the host already has an entry, it is replaced. Otherwise, a new entry is appended.
// It uses atomic write (temp file + rename) to avoid corruption from concurrent access.
func Store(host, fingerprint string) error {
	path, err := certFilePath()
	if err != nil {
		return err
	}

	// Read existing content
	existing, err := readLines(path)
	if err != nil {
		return err
	}

	// Remove any existing entry for this host
	var kept []string
	for _, line := range existing {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 && parts[0] == host {
			continue
		}
		kept = append(kept, line)
	}

	// Append the new entry
	kept = append(kept, fmt.Sprintf("%s %s", host, fingerprint))

	return atomicWriteLines(path, kept)
}

// Remove deletes all entries for the given host from ~/.uemcert.
// Returns true if any entries were removed.
func Remove(host string) (bool, error) {
	path, err := certFilePath()
	if err != nil {
		return false, err
	}

	lines, err := readLines(path)
	if err != nil {
		return false, err
	}

	var kept []string
	removed := false
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 && parts[0] == host {
			removed = true
			continue
		}
		kept = append(kept, line)
	}

	if !removed {
		return false, nil
	}

	return true, atomicWriteLines(path, kept)
}

// Entry represents a single pinned certificate entry.
type Entry struct {
	Host        string
	Fingerprint string
}

// List returns all pinned certificate entries from ~/.uemcert.
func List() ([]Entry, error) {
	path, err := certFilePath()
	if err != nil {
		return nil, err
	}

	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		entries = append(entries, Entry{Host: parts[0], Fingerprint: parts[1]})
	}
	return entries, nil
}

// readLines reads non-empty, non-comment lines from the cert file.
// Returns nil (not an error) if the file does not exist.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to open %s: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

// atomicWriteLines writes lines to a temp file and renames it into place.
func atomicWriteLines(path string, lines []string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".uemcert-tmp-*")
	if err != nil {
		return fmt.Errorf("unable to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Ensure cleanup on failure
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("unable to set permissions on temp file: %w", err)
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(tmp, line); err != nil {
			_ = tmp.Close()
			return fmt.Errorf("unable to write to temp file: %w", err)
		}
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("unable to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("unable to rename temp file: %w", err)
	}

	success = true
	return nil
}
