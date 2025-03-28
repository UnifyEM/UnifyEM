//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package uemservice

import (
	"time"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

type Service struct {
	logger         interfaces.Logger
	ServiceName    string
	ServiceVersion string
	ServiceBuild   int
	TaskTicker     time.Duration
	BackgroundFunc func(interfaces.Logger)
	TasksFunc      func(interfaces.Logger)
	StopFunc       func(interfaces.Logger)
	SEid           uint32
	tickerStop     chan struct{}
	tickerUpdate   chan time.Duration
}

// New returns a default Service
//
//goland:noinspection GoUnusedExportedFunction
func New(options ...func(*Service) error) (*Service, error) {

	// Initialize the Service with default values
	s := &Service{
		ServiceName:    "UEM",
		ServiceVersion: "unknown",
		TaskTicker:     60,
		SEid:           0,
		tickerStop:     make(chan struct{}),
		tickerUpdate:   make(chan time.Duration),
	}

	// Apply the options
	for _, op := range options {
		err := op(s)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Start the service
func (s *Service) Start() error {
	// run the os-specific function
	return s.start()
}

//goland:noinspection GoUnusedExportedFunction
func WithServiceName(name string) func(*Service) error {
	return func(s *Service) error {
		s.ServiceName = name
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithServiceVersion(version string) func(*Service) error {
	return func(s *Service) error {
		s.ServiceVersion = version
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithServiceBuild(build int) func(*Service) error {
	return func(s *Service) error {
		s.ServiceBuild = build
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithLogger(logger interfaces.Logger) func(*Service) error {
	return func(s *Service) error {
		s.logger = logger
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTaskTicker(ticker time.Duration) func(*Service) error {
	return func(s *Service) error {
		s.TaskTicker = ticker
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithBackgroundFunc(f func(interfaces.Logger)) func(*Service) error {
	return func(s *Service) error {
		s.BackgroundFunc = f
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithTasksFunc(f func(interfaces.Logger)) func(*Service) error {
	return func(s *Service) error {
		s.TasksFunc = f
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStopFunc(f func(interfaces.Logger)) func(*Service) error {
	return func(s *Service) error {
		s.StopFunc = f
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithSEid(seid uint32) func(*Service) error {
	return func(s *Service) error {
		s.SEid = seid
		return nil
	}
}

// UpdateTaskTicker updates the task ticker interval
func (s *Service) UpdateTaskTicker(newInterval time.Duration) {
	s.TaskTicker = newInterval
	s.tickerUpdate <- newInterval
}
