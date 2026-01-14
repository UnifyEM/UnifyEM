/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// TLSConfig returns a custom TLS configuration for the HTTP client
// In the case of a failure, it returns a default TLS configuration
// to avoid breaking the agent.
func (c *Communications) TLSConfig() *tls.Config {
	var err error

	// Load system root CA certificates
	roots, err := x509.SystemCertPool()
	if err != nil {
		return &tls.Config{}
	}

	// Create a custom TLS configuration
	// If the custom verification func returns a non-nil result, the handshake will fail
	return &tls.Config{
		RootCAs: roots,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {

			// Prevent a crash in the unlikely event that there are no verified chains
			if len(verifiedChains) == 0 {
				return fmt.Errorf("no verified chains found")
			}

			// If CA pinning is not enabled, return nil to accept the chain
			if !c.conf.AC.Get(schema.ConfigAgentPinCA).Bool() {
				return nil
			}

			// Get the CA hash from the configuration
			// If it is empty, we'll set it
			hash := c.conf.AP.Get(global.ConfigCAHash).String()

			// Iterate over each chain
			for _, chain := range verifiedChains {
				if len(chain) == 0 {
					continue
				}

				// Get the last certificate in the chain, which will be the CA
				cert := chain[len(chain)-1]

				// Compute the SHA-256 hash of the certificate's public key and base64 encode it
				pubKeyHash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
				pubKeyHashBase64 := base64.StdEncoding.EncodeToString(pubKeyHash[:])

				// If the hash is empty, set it in the configuration
				if hash == "" {
					c.conf.AP.Set(global.ConfigCAHash, pubKeyHashBase64)
					_ = c.conf.Checkpoint()
					return nil
				}

				// Accept the chain if the hash matches
				if pubKeyHashBase64 == hash {
					return nil
				}
			}

			// If no chain matches, return an error
			return fmt.Errorf("CA pinning error: no match")
		},
	}
}
