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

// Store appends a host and fingerprint entry to ~/.uemcert.
func Store(host, fingerprint string) error {
	path, err := certFilePath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("unable to open %s for writing: %w", path, err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s %s\n", host, fingerprint)
	if err != nil {
		return fmt.Errorf("unable to write to %s: %w", path, err)
	}
	return nil
}
