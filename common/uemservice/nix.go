//go:build !windows

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for operating systems other than windows

package uemservice

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnifyEM/UnifyEM/common/uemservice/privcheck"
)

// start the service
func (s *Service) start() error {
	if s.logger == nil {
		return errors.New("refusing to start service with nil logger")
	}

	root, err := privcheck.Check()
	if err != nil {
		s.logger.Errorf(s.SEid+4, "fatal error checking for admin privileges: %s", err.Error())
		return fmt.Errorf("fatal error checking for admin privileges: %w", err)
	}

	if !root {
		s.logger.Error(s.SEid+5, "service must be run with admin privileges", nil)
		return errors.New("service must be run with admin privileges")
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	s.logger.Infof(s.SEid+1, "%s %s (build %d) service started", s.ServiceName, s.ServiceVersion, s.ServiceBuild)
	s.logger.Debugf(s.SEid+1, "Debug logging enabled")

	if s.BackgroundFunc != nil {
		go s.BackgroundFunc(s.logger)
	}

	ticker := time.NewTicker(s.TaskTicker * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-s.tickerStop:
				return
			case newInterval := <-s.tickerUpdate:
				ticker.Stop()
				ticker = time.NewTicker(newInterval * time.Second)
			}
		}
	}()

	// Call the start function if defined
	if s.StartFunc != nil {
		s.StartFunc(s.logger)
	}

	// Loop, call the TasksFunc, and wait for an exit request
	for {
		select {
		case <-ticker.C:
			if s.TasksFunc != nil {
				s.TasksFunc(s.logger)
			}
		case <-signalChan:
			s.logger.Infof(s.SEid+2, "%s %s (build %d) service stopping", s.ServiceName, s.ServiceVersion, s.ServiceBuild)
			if s.StopFunc != nil {
				s.StopFunc(s.logger)
			}
			close(s.tickerStop)
			s.logger.Infof(s.SEid+3, "%s %s (build %d) service stopped", s.ServiceName, s.ServiceVersion, s.ServiceBuild)
			return nil
		}
	}
}
