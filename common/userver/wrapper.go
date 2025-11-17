/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package userver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/UnifyEM/UnifyEM/common/fields"
)

// ResponseWriterWrapper wraps a http.ResponseWriter to capture the status code
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *ResponseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Wrapper wraps a http.Handler to add standard headers, logging, and optionally authentication
func (s *HServer) Wrapper(handlerName string, h http.Handler, authFunc AuthFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Get the start time and source IP
		startTime := time.Now()
		src := s.getIP(req)

		// Check for authentication
		if authFunc != nil {
			authenticated, failMsg, details := authFunc(src, req.Header.Get("Authorization"))
			if !authenticated {
				s.Logger.Warning(s.SEid+12,
					"authentication failure",
					fields.NewFields(
						fields.NewField("src_ip", src),
						fields.NewField("method", req.Method),
						fields.NewField("uri", req.RequestURI),
						fields.NewField("handler", handlerName)))

				// Impose a time penalty for failed authentication
				s.PenaltyBox()

				// Return unauthorized status code
				w.WriteHeader(http.StatusUnauthorized)

				// If a failure message is provided, send it and ignore any errors
				if failMsg != nil {
					_, _ = w.Write(failMsg)
				}
				return
			}

			ctx := context.WithValue(req.Context(), "authDetails", details)
			req = req.WithContext(ctx)
		}

		// Create a context with a 30-second timeout
		ctx, cancel := context.WithTimeout(req.Context(), time.Duration(s.HandlerTimeout)*time.Second)
		defer cancel()

		// Create a new request with the timeout context
		req = req.WithContext(ctx)

		// Wrap the ResponseWriter to capture the status code
		rw := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Assume no timeout
		timeout := false

		// Call the actual handler to service the agent
		h.ServeHTTP(rw, req)

		// Check if the context timed out
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			timeout = true
		}

		// Set requested reply headers
		for _, header := range s.Headers {
			w.Header().Set(header.Key, header.Value)
		}

		// Get duration of agent
		duration := time.Since(startTime)

		// Remove parameters from URI to avoid logging confidential information
		uri := strings.Split(req.RequestURI, "?")[0]

		logFields := fields.NewFields(
			fields.NewField("code", rw.statusCode),
			fields.NewField("src_ip", src),
			fields.NewField("method", req.Method),
			fields.NewField("uri", uri),
			fields.NewField("handler", handlerName),
			fields.NewField("duration", fmt.Sprintf("%.4f", duration.Seconds())))

		if timeout {
			logFields.Append(fields.NewField("timeout", "true"))
		}

		// Log the event
		s.Logger.Info(s.SEid+10, "HTTP", logFields)
	})
}

// getIP returns an IP address by reading the forwarded-for
// header (for proxies or load balancers) and falls back to use the remote address.
func (s *HServer) getIP(r *http.Request) string {
	var source = ""
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		source = forwarded
	} else {
		source = r.RemoteAddr
	}

	// Clean up, remove port number
	if len(source) > 0 {
		if strings.HasPrefix(source, "[") {
			// IPv6 address
			t := strings.Split(source, "]")
			source = t[0][1:]
		} else {
			// IPv4 - hack off port number
			t := strings.Split(source, ":")
			source = t[0]
		}
	}
	return source
}
