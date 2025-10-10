//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Code for operating systems other than windows
//go:build windows

package uemservice

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
)

// start starts the service.
func (s *Service) start() error {

	// Run the Windows Service
	err := svc.Run(s.ServiceName, s)
	if err != nil {
		return fmt.Errorf("%s service failed: %v", s.ServiceName, err)
	}
	return nil
}

// Execute runs the Windows service
//
//goland:noinspection GoUnusedParameter
func (s *Service) Execute(args []string, req <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	status <- svc.Status{State: svc.StartPending}

	if s.logger == nil {
		return false, 1
	}

	s.logger.Infof(s.SEid+1, "%s %s (build %d) service started", s.ServiceName, s.ServiceVersion, s.ServiceBuild)
	s.logger.Debugf(s.SEid+1, "Debug logging enabled")

	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

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

loop:
	for {
		select {
		case c := <-req:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				s.logger.Infof(s.SEid+2, "%s %s (build %d) service stopping", s.ServiceName, s.ServiceVersion, s.ServiceBuild)
				if s.StopFunc != nil {
					s.StopFunc(s.logger)
				}
				close(s.tickerStop)
				break loop
			default:
			}
		case <-ticker.C:
			if s.TasksFunc != nil {
				s.TasksFunc(s.logger)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	s.logger.Infof(s.SEid+3, "%s %s (build %d) service stopped", s.ServiceName, s.ServiceVersion, s.ServiceBuild)
	return false, 0
}
