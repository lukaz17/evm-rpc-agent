// Copyright (C) 2025 Nguyen Nhat Tung
// This file is part of EVM RPC Agent
//
// EVM RPC Agent is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// EVM RPC Agent is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with EVM RPC Agent. If not, see <https://www.gnu.org/licenses/>.

package svc

import (
	"sync"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// Scheduler host and dispatch recurring jobs at scheduled intervals.
type Scheduler struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal

	jobs     map[string]*ScheduledJob
	jobsMu   sync.RWMutex
	tickMs   int64
	exitChan chan struct{}
	wg       sync.WaitGroup
}

// Return new Scheduler instance.
func NewScheduler(tickMs int64, logger zerolog.Logger) *Scheduler {
	if tickMs < 250 {
		tickMs = 250
	}
	if tickMs > 60000 {
		tickMs = 60000
	}

	svc := &Scheduler{
		jobs:     make(map[string]*ScheduledJob),
		tickMs:   tickMs,
		exitChan: make(chan struct{}),
	}
	svc.i = svc.InitServiceCore("Scheduler", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

func (s *Scheduler) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "schedule_job":
		p := NewSchedulerParams(msg)

		s.jobsMu.Lock()
		s.jobs[p.JobID] = &ScheduledJob{
			IntervalMs:      p.IntervalMs,
			ServiceID:       p.ServiceID,
			Command:         p.Command,
			Params:          p.Params,
			NextExecutionMs: time.Now().UnixMilli(),
		}
		s.jobsMu.Unlock()

		if s.WorkerCount() == 0 {
			s.SetWorker(uint64(1))
		}

		s.i.Logger.Infof("%s#%02d: Job %q scheduled (interval=%dms, service=%s, cmd=%s).",
			s.i.ServiceID, workerID, p.JobID, p.IntervalMs, p.ServiceID, p.Command)
		msg.Return(true)

	case "unschedule_job":
		p := NewSchedulerParams(msg)
		s.jobsMu.Lock()
		delete(s.jobs, p.JobID)
		s.jobsMu.Unlock()
		s.i.Logger.Infof("%s#%02d: Job %q unscheduled.", s.i.ServiceID, workerID, p.JobID)
		msg.Return(true)

	case "exit":
		close(s.exitChan)
		s.wg.Wait()
		s.i.Logger.Infof("%s#%02d: Scheduler stopped.", s.i.ServiceID, workerID)
		msg.Return(true)

	default:
		s.i.Logger.Warnf("%s#%02d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *Scheduler) SetWorker(workerCount uint64) {
	wasZero := s.WorkerCount() == 0
	s.ServiceCore.SetWorker(workerCount)
	if workerCount > 0 && wasZero {
		s.wg.Add(1)
		go s.tickLoop()
	}
}

func (s *Scheduler) tickLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(time.Duration(s.tickMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.exitChan:
			return
		case <-ticker.C:
			s.executeDueJobs()
		}
	}
}

func (s *Scheduler) executeDueJobs() {
	now := time.Now().UnixMilli()

	s.jobsMu.RLock()
	dueJobIDs := make([]string, 0, len(s.jobs))
	for id, job := range s.jobs {
		if job.NextExecutionMs <= now {
			dueJobIDs = append(dueJobIDs, id)
		}
	}
	s.jobsMu.RUnlock()

	for _, id := range dueJobIDs {
		s.jobsMu.RLock()
		job := s.jobs[id]
		s.jobsMu.RUnlock()
		if job != nil {
			s.Dispatch(job.ServiceID, job.Command, job.Params)
		}
	}

	if len(dueJobIDs) > 0 {
		s.jobsMu.Lock()
		for _, id := range dueJobIDs {
			job := s.jobs[id]
			if job != nil && job.IntervalMs > 0 {
				job.NextExecutionMs = now - (now % job.IntervalMs) + job.IntervalMs
			} else if job != nil && job.IntervalMs <= 0 {
				delete(s.jobs, id)
			}
		}
		s.jobsMu.Unlock()
	}
}

// SchedulerParams contain the parameters for scheduling and unscheduling jobs.
type SchedulerParams struct {
	Command    string
	IntervalMs int64
	JobID      string
	Params     multiplex.ExecParams
	ServiceID  string
}

// NewSchedulerParams extracts scheduling parameters from a service message.
func NewSchedulerParams(msg *multiplex.ServiceMessage) *SchedulerParams {
	return &SchedulerParams{
		Command:    msg.GetParam("command", "").(string),
		IntervalMs: msg.GetParam("interval_ms", int64(0)).(int64),
		JobID:      msg.GetParam("job_id", "").(string),
		Params:     msg.GetParam("params", multiplex.ExecParams{}).(multiplex.ExecParams),
		ServiceID:  msg.GetParam("service_id", "").(string),
	}
}

// ScheduledJob contain parameters for a recurring job.
type ScheduledJob struct {
	IntervalMs      int64
	ServiceID       string
	Command         string
	Params          multiplex.ExecParams
	NextExecutionMs int64
}
