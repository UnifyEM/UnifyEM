/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"fmt"
	"time"

	"go.etcd.io/bbolt"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// A separate package with a struct are used for looser coupling with the database

type DB struct {
	db     *bbolt.DB
	logger interfaces.Logger
}

const BucketAuth = "Auth"
const BucketAgentRequests = "Requests"
const BucketAgentMeta = "AgentMeta"
const BucketAgentEvents = "AgentEvents"
const BucketUserMeta = "UserMeta"

var bucketList = []string{BucketAuth, BucketAgentRequests, BucketAgentMeta, BucketAgentEvents, BucketUserMeta}

// Open opens (or creates) a Bolt DB at the specified path.
// It also creates three buckets if they do not already exist.
func Open(filePath string, logger interfaces.Logger) (*DB, error) {

	logger.Infof(2201, "Opening database: %s", filePath)
	// Open the Bolt DB file. 0600 means read/write permissions for the current user only.
	// The Timeout option allows Bolt to wait if the file is locked by another process.
	db, err := bbolt.Open(filePath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt db: %w", err)
	}

	// Create all buckets within a single transaction if they don't already exist.
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range bucketList {
			_, createErr := tx.CreateBucketIfNotExists([]byte(bucketName))
			if createErr != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, createErr)
			}
		}
		return nil
	})
	if err != nil {
		// If creating buckets failed, close the DB to avoid resource leaks.
		_ = db.Close()
		return nil, err
	}

	return &DB{db: db}, nil
}

// Close the database, ignore any errors
func (d *DB) Close() {
	_ = d.db.Close()
}
