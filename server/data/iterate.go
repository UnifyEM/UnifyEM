/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

// ForEach iterates over all keys in the specified bucket and applies the given function
func (d *Data) ForEach(bucketName string, fn func(key, value []byte) error) error {
	return d.database.ForEach(bucketName, func(key, value []byte) error {
		return fn(key, value)
	})
}
