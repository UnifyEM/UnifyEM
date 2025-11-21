# Elliptic Curve Cryptography Implementation Plan

## Overview

This document describes the implementation of elliptic curve cryptography for the UnifyEM project. The system uses P-384 elliptic curves with dual keypairs (one for signing, one for encryption) for both server and agents.

## Cryptographic Approach

### Key Generation
- **Curve**: P-384 (NIST standard)
- **Keypairs**: Two per entity (server/agent)
  - Signature keypair (for ECDSA signatures)
  - Encryption keypair (for hybrid encryption)
- **Format**: All keys stored and transmitted as base64-encoded strings

### Encryption Scheme
- **Method**: Hybrid encryption
  1. Generate random AES-GCM key
  2. Encrypt data with AES-GCM key
  3. Encrypt AES-GCM key with recipient's P-384 public encryption key
  4. Return combined encrypted payload (base64-encoded)

### Signature Scheme
- **Method**: ECDSA with SHA-256
- Uses private signature key to sign
- Uses public signature key to verify

## Implementation Components

### 1. Crypto Functions (`common/crypto/`)

#### Files and Functions

**`generate.go`**
```go
GenerateKeyPairs() (privateSig, publicSig, privateEnc, publicEnc string, err error)
```
- Generates two P-384 keypairs
- Returns all keys as base64-encoded strings

**`encrypt.go`**
```go
Encrypt(data []byte, recipientPublicEncKey string) (encrypted string, err error)
```
- Encrypts data using hybrid encryption
- Returns base64-encoded encrypted payload

**`decrypt.go`**
```go
Decrypt(encrypted string, privateEncKey string) (data []byte, err error)
```
- Decrypts payload using private encryption key
- Returns original data

**`sign.go`**
```go
Sign(data []byte, privateSignKey string) (signature string, err error)
```
- Signs data using ECDSA with SHA-256
- Returns base64-encoded signature

**`verify.go`**
```go
Verify(data []byte, signature string, publicSignKey string) (bool, error)
```
- Verifies signature using public signature key
- Returns true if valid, false otherwise

### 2. Configuration Storage

#### Agent Configuration (`agent/global/defaults.go`)

**Config Set**: `ap` (agent persistent)

**Fields** (6 total):
- `server_public_sig` - Server's public signature key
- `server_public_enc` - Server's public encryption key
- `ec_private_sig` - Agent's private signature key
- `ec_public_sig` - Agent's public signature key
- `ec_private_enc` - Agent's private encryption key
- `ec_public_enc` - Agent's public encryption key

#### Server Configuration (`server/global/defaults.go`)

**Config Set**: `sp` (server persistent)

**Fields** (4 total):
- `ec_private_sig` - Server's private signature key
- `ec_public_sig` - Server's public signature key
- `ec_private_enc` - Server's private encryption key
- `ec_public_enc` - Server's public encryption key

### 3. Schema Changes

#### Agent Data Structure
Add fields to store client's public keys:
```go
ClientPublicSig string `json:"client_public_sig,omitempty"`
ClientPublicEnc string `json:"client_public_enc,omitempty"`
```

#### Registration Request
Agent sends its public keys:
```go
ClientPublicSig string `json:"client_public_sig,omitempty"`
ClientPublicEnc string `json:"client_public_enc,omitempty"`
```

#### Registration Response
Server sends its public keys:
```go
ServerPublicSig string `json:"server_public_sig,omitempty"`
ServerPublicEnc string `json:"server_public_enc,omitempty"`
```

#### Token Request
Agent sends its public keys with every token request:
```go
ClientPublicSig string `json:"client_public_sig,omitempty"`
ClientPublicEnc string `json:"client_public_enc,omitempty"`
```

#### Token Response
Server sends its public keys with every token response:
```go
ServerPublicSig string `json:"server_public_sig,omitempty"`
ServerPublicEnc string `json:"server_public_enc,omitempty"`
```

## Key Exchange Flow

### Initial Agent Installation

1. Agent generates two keypairs during installation
2. Agent stores 4 keys (2 private, 2 public) in AP config
3. During registration:
   - Agent sends its 2 public keys to server
   - Server stores agent's public keys in database
   - Server sends its 2 public keys to agent
   - Agent stores server's public keys in AP config

### Existing Agents (Legacy Support)

1. Server generates keypairs on startup if missing
2. Agent generates keypairs on startup if missing
3. During token refresh:
   - Agent sends its 2 public keys (if generated)
   - Server stores them if it doesn't have them
   - Server sends its 2 public keys
   - Agent stores them if it doesn't have them

### Server Startup

1. Check SP config for 4 keys
2. If any missing, generate both keypairs
3. Store all 4 keys in SP config

### Agent Startup

1. Check AP config for 4 keys (agent's own)
2. If any missing, generate both keypairs
3. Store all 4 keys in AP config

## Key Management Rules

### Server Behavior

**On receiving client public keys:**
- If server has no keys for this agent: Store them
- If server has different keys: Log WARNING, keep existing keys (don't replace)

**On sending server public keys:**
- Always include server's 2 public keys in token responses

### Agent Behavior

**On receiving server public keys:**
- If agent has no server keys: Store them
- If agent has different server keys: Log WARNING, keep existing keys (don't replace)

**On sending client public keys:**
- Always include agent's 2 public keys in token requests

### Rekey Operation

The `rekey` command should:
1. Delete server's public keys from agent config
2. Force agent to re-fetch server public keys on next token request
3. Optionally: Generate new agent keypairs (TBD)

## Security Considerations

1. **Private keys never leave their origin**
   - Agent private keys stay on agent
   - Server private keys stay on server

2. **Public key pinning**
   - Once stored, public keys are not replaced
   - Prevents MITM attacks
   - Requires manual rekey if server keys change

3. **Hybrid encryption**
   - Combines security of EC with performance of AES
   - Each encryption uses fresh random AES key

4. **Key size**
   - P-384 provides ~192-bit security level
   - Suitable for long-term data protection

## Usage (Future)

The encryption/signing capabilities will be gradually adopted:
- Phase 1: Key infrastructure deployment (this implementation)
- Phase 2: Encrypt sensitive command parameters
- Phase 3: Sign critical requests for integrity
- Phase 4: Additional use cases as identified

## Backward Compatibility

- Encryption is not mandatory initially
- Agents without keys will generate them on startup
- Server without keys will generate them on startup
- Existing communication continues to work
- Encrypted features will be added incrementally
