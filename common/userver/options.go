//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package userver

import "github.com/UnifyEM/UnifyEM/common/interfaces"

// Functional options

//goland:noinspection GoUnusedExportedFunction
func WithLogger(logger interfaces.Logger) func(*HServer) error {
	return func(e *HServer) error {
		e.Logger = logger
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithListen(listen string) func(*HServer) error {
	return func(e *HServer) error {
		e.Listen = listen
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHTTPTimeout(t int) func(*HServer) error {
	return func(e *HServer) error {
		e.HTTPTimeout = t
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHTTPIdleTimeout(t int) func(*HServer) error {
	return func(e *HServer) error {
		e.HTTPIdleTimeout = t
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHandlerTimeout(t int) func(*HServer) error {
	return func(e *HServer) error {
		e.HandlerTimeout = t
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithPenaltyBox(min, max int) func(*HServer) error {
	return func(e *HServer) error {
		e.PenaltyBoxMin = min
		e.PenaltyBoxMax = max
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithMaxConcurrent(m int) func(*HServer) error {
	return func(e *HServer) error {
		e.MaxConcurrent = m
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithLogFile(logfile string) func(*HServer) error {
	return func(e *HServer) error {
		e.LogFile = logfile
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithDownFile(down string) func(*HServer) error {
	return func(e *HServer) error {
		e.DownFile = down
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithSEid(seid uint32) func(*HServer) error {
	return func(e *HServer) error {
		e.SEid = seid
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHealthHandler(h bool) func(*HServer) error {
	return func(e *HServer) error {
		e.HealthHandler = h
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTestHandler(t bool) func(*HServer) error {
	return func(e *HServer) error {
		e.TestHandler = t
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStrictSlash(s bool) func(*HServer) error {
	return func(e *HServer) error {
		e.StrictSlash = s
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithDefaultHeaders(d bool) func(*HServer) error {
	return func(e *HServer) error {
		e.DefaultHeaders = d
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTLS(t bool) func(*HServer) error {
	return func(e *HServer) error {
		e.TLS = t
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTLSCertFile(certFile string) func(*HServer) error {
	return func(e *HServer) error {
		e.TLSCertFile = certFile
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTLSKeyFile(keyFile string) func(*HServer) error {
	return func(e *HServer) error {
		e.TLSKeyFile = keyFile
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTLSStrongCiphers(c bool) func(*HServer) error {
	return func(e *HServer) error {
		e.TLSStrongCiphers = c
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithDebug(d bool) func(*HServer) error {
	return func(e *HServer) error {
		e.Debug = d
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithFileDir(pattern, dir string, authFunc AuthFunc) func(*HServer) error {
	return func(e *HServer) error {
		e.FileSrv.Dir = dir
		e.FileSrv.Pattern = pattern
		e.FileSrv.AuthFunc = authFunc
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithAuthFunc(authFunc AuthFunc) func(*HServer) error {
	return func(e *HServer) error {
		e.AuthFunc = authFunc
		return nil
	}
}
