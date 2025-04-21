//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package null

import (
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// LoggerNull implements interfaces.Logger and discards all data
// This is useful when a function requires a logger but is being called
// for testing for from the command line where a logger is not needed or
// readily available.
type LoggerNull struct{}

func Logger() interfaces.Logger {
	return &LoggerNull{}
}

func (n *LoggerNull) Debug(_ uint32, _ string, _ interfaces.Fields) {
}

func (n *LoggerNull) Info(_ uint32, _ string, _ interfaces.Fields) {
}

func (n *LoggerNull) Warning(_ uint32, _ string, _ interfaces.Fields) {
}

func (n *LoggerNull) Error(_ uint32, _ string, _ interfaces.Fields) {
}

func (n *LoggerNull) Fatal(_ uint32, _ string, _ interfaces.Fields) {
}

func (n *LoggerNull) Debugf(_ uint32, _ string, _ ...any) {
}

func (n *LoggerNull) Infof(_ uint32, _ string, _ ...any) {
}

func (n *LoggerNull) Warningf(_ uint32, _ string, _ ...any) {
}

func (n *LoggerNull) Errorf(_ uint32, _ string, _ ...any) {
}

func (n *LoggerNull) Fatalf(_ uint32, _ string, _ ...any) {
}
