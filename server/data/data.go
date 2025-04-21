//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package data

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/hasher"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/server/db"
	"github.com/UnifyEM/UnifyEM/server/global"
)

type Data struct {
	logger            interfaces.Logger
	conf              *global.ServerConfig
	database          *db.DB
	hasher            *hasher.Hasher
	jwtKey            []byte
	BucketAuth        string
	BucketRequests    string
	BucketAgentMeta   string
	BucketAgentStatus string
}

// New creates a new Data instance
func New(conf *global.ServerConfig, logger interfaces.Logger) (*Data, error) {
	var err error

	// Get or create the JWTKey
	jwtKey := conf.SP.Get(global.ConfigJWTKey).Bytes()
	if len(jwtKey) == 0 {

		// Generate a new key
		jwtKey, err = randomBytes(global.TokenLength)
		if err != nil {
			return nil, fmt.Errorf("unable to generate JWT key: %w", err)
		}

		// Save the key to the configuration
		conf.SP.Set(global.ConfigJWTKey, jwtKey)
	}

	// Get database path. If it doesn't exist, it will be created by global.Config()
	dbPath := conf.SC.Get(global.ConfigDBPath).String()
	if dbPath == "" {
		return nil, errors.New("database path missing from configuration")
	}

	dbInstance, err := db.Open(filepath.Join(dbPath, strings.ToLower(global.Name)+".db"), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to open or create database: %w", err)
	}

	return &Data{
		logger:          logger,
		conf:            conf,
		database:        dbInstance,
		jwtKey:          jwtKey,
		hasher:          hasher.New(hasher.WithCache(global.MemoryCacheTTL)),
		BucketAuth:      db.BucketAuth,
		BucketRequests:  db.BucketAgentRequests,
		BucketAgentMeta: db.BucketAgentMeta,
	}, nil
}

// Close anything data-related that requires it.
func (d *Data) Close() {

	// If the data instance is nil, bail
	if d == nil {
		return
	}

	// Close the database connection
	if d.database != nil {
		d.database.Close()
	}
}
