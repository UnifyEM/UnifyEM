/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package userver

import (
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

type HServer struct {
	Headers          Headers
	Routes           Routes
	Listen           string
	HTTPTimeout      int
	HTTPIdleTimeout  int
	HandlerTimeout   int
	MaxConcurrent    int
	PenaltyBoxMin    int
	PenaltyBoxMax    int
	LogFile          string // Optional, defaults to stdout
	DownFile         string
	HealthHandler    bool
	TestHandler      bool
	StrictSlash      bool
	DefaultHeaders   bool
	TLS              bool
	TLSCertFile      string
	TLSKeyFile       string
	TLSStrongCiphers bool
	Debug            bool
	AuthFunc         AuthFunc // Used for not found and method not allowed handlers
	server           *http.Server
	Logger           interfaces.Logger
	SEid             uint32 // Starting event ID for logging
	FileSrv          FileServer
}

type FileServer struct {
	Dir      string
	Pattern  string
	AuthFunc AuthFunc
}

// AuthFunc is used as a callback to authenticate requests
// It returns a bool to indicate success or failure
// In the event of a failure, []byte may contain a message to send
// The "any" type is passed through to the handler in the context
type AuthFunc func(string, string) (bool, []byte, any)

// AuthDetails is an interface that should be implemented by the
// application to provide details about the authenticated user
type AuthDetails interface {
	IsAuthenticated() bool
}

// Route defines a route for the HTTP router. It can include a
// standard handler that returns a http.Handler or a JHandler
// that returns a JResponse structure.
type Route struct {
	Name     string
	Methods  []string
	Pattern  string
	Handler  http.Handler
	JHandler JHandler
	AuthFunc AuthFunc
}

type Routes []Route

type Header struct {
	Key   string
	Value string
}

type Headers []Header

// Response provides a consistent set of fields for API responses
type Response struct {
	Status  string `json:"status"`            // Text Status
	Code    int    `json:"code"`              // HTTP status code
	Details string `json:"details,omitempty"` // optional response details
	Data    any    `json:"data,omitempty"`    // any type of data
}

// JHandler is the type of the function to be wrapped
type JHandler func(req *http.Request) JResponse

// JResponse is the structure returned by the wrapped function
type JResponse struct {
	HTTPCode int
	JSONData any
}
