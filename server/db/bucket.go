//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package db

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.etcd.io/bbolt"
)

// SetData serializes and stores data in a specified bucket using a given key
func (d *DB) SetData(bucketName string, key string, value interface{}) error {

	// Serialize the value
	data, err := d.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	// Store the serialized data in the bucket
	err = d.db.Update(func(tx *bbolt.Tx) error {

		// Get or create the specified bucket
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("%s bucket not found", bucketName)
		}

		// Store the serialized data in the bucket with the given key
		err = bucket.Put([]byte(key), data)
		if err != nil {
			return fmt.Errorf("failed to store data in bucket: %w", err)
		}

		return nil
	})

	return err
}

// GetData retrieves and deserializes data from a specified bucket using a given key
func (d *DB) GetData(bucketName string, key string, result interface{}) error {
	err := d.db.View(func(tx *bbolt.Tx) error {

		// Get the specified bucket
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		// Retrieve the serialized data for the given key
		data := bucket.Get([]byte(key))
		if data == nil {
			return errors.New("key not found")
		}

		// If passed a nil result, don't deserialize the data
		if result != nil {
			// Deserialize the stored data into the result
			err := d.deserialize(data, result)
			if err != nil {
				return fmt.Errorf("failed to deserialize data: %w", err)
			}
		}

		return nil
	})

	return err
}

// DeleteData deletes data from a specified bucket using a given key
func (d *DB) DeleteData(bucketName string, key string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {

		// Get the specified bucket
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		/* This seems like an unnecessary database hit
		// Check if the key exists in the bucket
		if bucket.Get([]byte(key)) == nil {
			return errors.New("key not found")
		}
		*/

		// Delete the entry for the given key
		if err := bucket.Delete([]byte(key)); err != nil {
			return fmt.Errorf("error deleting data %w", err)
		}
		return nil
	})
}

// KeyExists checks if a key exists in a specified bucket
func (d *DB) KeyExists(bucketName string, key string) (bool, error) {
	var exists bool
	err := d.db.View(func(tx *bbolt.Tx) error {
		// Get the specified bucket
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		// Check if the key exists in the bucket
		exists = bucket.Get([]byte(key)) != nil

		return nil
	})
	return exists, err
}

// ForEach iterates over all keys in the specified bucket and applies the given function
func (d *DB) ForEach(bucketName string, fn func(key, value []byte) error) error {
	return d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.ForEach(fn)
	})
}

// serialize converts a struct into a byte slice
func (d *DB) serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// deserialize converts the stored data into a struct
func (d *DB) deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
