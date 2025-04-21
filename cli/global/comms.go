//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package global

import "github.com/UnifyEM/UnifyEM/cli/util"

type Comms interface {
	SetToken(token string)
	Post(endpoint string, payload interface{}) (int, []byte, error)
	Put(endpoint string, payload interface{}) (int, []byte, error)
	Get(endpoint string) (int, []byte, error)
	GetQuery(endpoint string, pairs *util.NVPairs) (int, []byte, error)
	Delete(endpoint string) (int, []byte, error)
}
