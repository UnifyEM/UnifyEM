/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/UnifyEM/UnifyEM/cli/certstore"
	"github.com/UnifyEM/UnifyEM/cli/global"
)

// UntrustedCertError is returned when a server presents a certificate
// that is neither trusted by the system roots nor stored in ~/.uemcert.
type UntrustedCertError struct {
	Cert *x509.Certificate
	Host string
}

func (e *UntrustedCertError) Error() string {
	return fmt.Sprintf("untrusted certificate from %s", e.Host)
}

// sendRequest is a lower level function that sends HTTP requests
func (c *Communications) sendRequest(method, endpoint string, payload []byte) (int, []byte, error) {

	// Build the request URL
	reqURL := fmt.Sprintf("%s%s", global.ServerURL, endpoint)

	// Extract host:port for certificate operations
	host := hostFromURL(reqURL)

	// Build the HTTP client with TLS certificate verification
	client := c.buildHTTPClient(host)

	code, body, err := c.doRequest(client, method, reqURL, payload)
	if err != nil {
		// Check if this is an untrusted certificate error
		var certErr *UntrustedCertError
		if errors.As(err, &certErr) {
			// Prompt the user to accept the certificate
			accepted, promptErr := promptUserForCert(certErr.Cert, certErr.Host)
			if promptErr != nil {
				return 0, nil, fmt.Errorf("failed to prompt for certificate acceptance: %w", promptErr)
			}
			if !accepted {
				return 0, nil, fmt.Errorf("certificate rejected by user")
			}

			// Store the accepted fingerprint
			fp := certstore.Fingerprint(certErr.Cert)
			if storeErr := certstore.Store(host, fp); storeErr != nil {
				return 0, nil, fmt.Errorf("failed to store certificate: %w", storeErr)
			}

			// Retry with the now-trusted certificate
			client = c.buildHTTPClient(host)
			return c.doRequest(client, method, reqURL, payload)
		}
		return 0, nil, err
	}
	return code, body, nil
}

// doRequest executes an HTTP request and returns status code, body, and error.
func (c *Communications) doRequest(client *http.Client, method, reqURL string, payload []byte) (int, []byte, error) {

	httpReq, err := http.NewRequest(method, reqURL, bytes.NewBuffer(payload))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the Authorization header if a token is present
	if c.token != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	// Set the appropriate headers
	if method == "POST" || method == "PUT" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response body
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, responseBody.Bytes(), nil
}

// buildHTTPClient creates an http.Client with custom TLS verification
// that trusts certificates whose fingerprints are stored in ~/.uemcert.
func (c *Communications) buildHTTPClient(host string) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("server presented no certificates")
			}

			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("failed to parse server certificate: %w", err)
			}

			// First, try standard system root verification
			roots, err := x509.SystemCertPool()
			if err == nil && roots != nil {
				// Extract hostname without port for DNSName verification
				dnsName, _, _ := strings.Cut(host, ":")
				opts := x509.VerifyOptions{
					DNSName:       dnsName,
					Roots:         roots,
					Intermediates: x509.NewCertPool(),
				}
				// Add intermediate certs if present
				for _, rawCert := range rawCerts[1:] {
					intermediateCert, parseErr := x509.ParseCertificate(rawCert)
					if parseErr == nil {
						opts.Intermediates.AddCert(intermediateCert)
					}
				}
				if _, verifyErr := cert.Verify(opts); verifyErr == nil {
					return nil
				}
			}

			// System verification failed — check the fingerprint store
			fp := certstore.Fingerprint(cert)
			trusted, err := certstore.IsTrusted(host, fp)
			if err != nil {
				return fmt.Errorf("failed to check certificate store: %w", err)
			}
			if trusted {
				return nil
			}

			// Not trusted by any means — return a typed error for prompting
			return &UntrustedCertError{Cert: cert, Host: host}
		},
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

// promptUserForCert displays certificate details and asks the user to accept or reject.
func promptUserForCert(cert *x509.Certificate, host string) (bool, error) {
	fmt.Printf("\nWARNING: The server at %s presented an untrusted certificate:\n\n", host)
	fmt.Print(certstore.FormatCertDetails(cert))
	fmt.Print("\nDo you want to trust this certificate? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "yes" || answer == "y", nil
}

// hostFromURL extracts the host:port from a URL string.
// If no port is specified, defaults based on scheme.
func hostFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return host + ":" + port
}
