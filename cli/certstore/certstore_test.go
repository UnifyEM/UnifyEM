/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package certstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// generateTestCert creates a self-signed certificate for testing.
func generateTestCert(t *testing.T) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   "test.example.com",
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		DNSNames:  []string{"test.example.com", "localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	return cert
}

func TestFingerprint(t *testing.T) {
	cert := generateTestCert(t)
	fp := Fingerprint(cert)
	if fp == "" {
		t.Fatal("fingerprint should not be empty")
	}
	if len(fp) != 64 {
		t.Fatalf("expected 64 hex characters, got %d", len(fp))
	}

	// Same cert should produce same fingerprint
	fp2 := Fingerprint(cert)
	if fp != fp2 {
		t.Fatal("fingerprint should be deterministic")
	}
}

func TestFormatCertDetails(t *testing.T) {
	cert := generateTestCert(t)
	details := FormatCertDetails(cert)
	if details == "" {
		t.Fatal("cert details should not be empty")
	}

	// Check that key fields are present
	for _, expected := range []string{"Subject:", "Issuer:", "Fingerprint:", "Not Before:", "Not After:", "DNS Names:"} {
		if !strings.Contains(details, expected) {
			t.Errorf("expected %q in cert details", expected)
		}
	}
}

func TestStoreAndIsTrusted(t *testing.T) {
	// Use a temp directory as home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	host := "server.example.com:443"
	fp := "AABBCCDD0011223344556677889900AABBCCDD0011223344556677889900AABB"

	// Should not be trusted initially
	trusted, err := IsTrusted(host, fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trusted {
		t.Fatal("should not be trusted before storing")
	}

	// Store it
	err = Store(host, fp)
	if err != nil {
		t.Fatalf("unexpected error storing: %v", err)
	}

	// Should now be trusted
	trusted, err = IsTrusted(host, fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !trusted {
		t.Fatal("should be trusted after storing")
	}

	// Different host should not be trusted
	trusted, err = IsTrusted("other.example.com:443", fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trusted {
		t.Fatal("different host should not be trusted")
	}

	// Different fingerprint should not be trusted
	trusted, err = IsTrusted(host, "DIFFERENT_FINGERPRINT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trusted {
		t.Fatal("different fingerprint should not be trusted")
	}

	// Verify file permissions
	path := filepath.Join(tmpDir, certFile)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("unable to stat cert file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected file permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestIsTrustedMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	trusted, err := IsTrusted("host:443", "FP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trusted {
		t.Fatal("should not be trusted when file does not exist")
	}
}

func TestStoreMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	entries := []struct {
		host string
		fp   string
	}{
		{"host1:443", "FP1"},
		{"host2:8443", "FP2"},
		{"host3:443", "FP3"},
	}

	for _, e := range entries {
		if err := Store(e.host, e.fp); err != nil {
			t.Fatalf("unexpected error storing %s: %v", e.host, err)
		}
	}

	for _, e := range entries {
		trusted, err := IsTrusted(e.host, e.fp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !trusted {
			t.Fatalf("%s should be trusted", e.host)
		}
	}
}

