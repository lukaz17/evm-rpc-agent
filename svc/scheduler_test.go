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
	"testing"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

func TestScheduler_ScheduleJob(t *testing.T) {
	ctrl := newTestControllerWithTickMs(t, 250, nil)
	msg := multiplex.ExecParams{
		"job_id":      "test_schedule",
		"interval_ms": int64(1000),
		"service_id":  "Notify",
		"command":     "last_update",
	}
	ctrl.DispatchOnce("Scheduler", "schedule_job", msg)

	result := msg.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}
}

func TestScheduler_UnscheduleJob(t *testing.T) {
	ctrl := newTestControllerWithTickMs(t, 250, nil)

	scheduleMsg := multiplex.ExecParams{
		"job_id":      "test_unschedule",
		"interval_ms": int64(1000),
		"service_id":  "Notify",
		"command":     "last_update",
	}
	ctrl.Dispatch("Scheduler", "schedule_job", scheduleMsg)

	unscheduleMsg := multiplex.ExecParams{
		"job_id": "test_unschedule",
	}
	ctrl.DispatchOnce("Scheduler", "unschedule_job", unscheduleMsg)

	result := unscheduleMsg.ReturnResult()
	if result != true {
		t.Fatalf("expected return value true, got %v", result)
	}
}

func TestScheduler_UnknownCommand(t *testing.T) {
	ctrl := newTestControllerWithTickMs(t, 250, nil)
	msg := multiplex.ExecParams{}
	ctrl.DispatchOnce("Scheduler", "unknown_cmd", msg)

	result := msg.ReturnResult()
	if result != nil {
		t.Fatalf("expected return value nil, got %v", result)
	}
}

func TestScheduler_NewScheduler_TickMsClamp_Min(t *testing.T) {
	logger := zerolog.Nop()
	scheduler := NewScheduler(100, logger)
	if scheduler.tickMs != 250 {
		t.Fatalf("tickMs = %d, want 250", scheduler.tickMs)
	}
}

func TestScheduler_NewScheduler_TickMsClamp_Max(t *testing.T) {
	logger := zerolog.Nop()
	scheduler := NewScheduler(100000, logger)
	if scheduler.tickMs != 60000 {
		t.Fatalf("tickMs = %d, want 60000", scheduler.tickMs)
	}
}

func TestScheduler_NewScheduler_TickMs_Normal(t *testing.T) {
	logger := zerolog.Nop()
	scheduler := NewScheduler(1000, logger)
	if scheduler.tickMs != 1000 {
		t.Fatalf("tickMs = %d, want 1000", scheduler.tickMs)
	}
}

func TestScheduler_ExecutionFlow(t *testing.T) {
	mem.LastUpdate = 0
	ctrl := newTestControllerWithTickMs(t, 250, nil)
	t.Cleanup(func() { ctrl.svc.Exec("exit", multiplex.ExecParams{}) })

	ctrl.ScheduleJob("test_exec_flow", 500, "Notify", "last_update", multiplex.ExecParams{})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if mem.LastUpdate > 0 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatal("LastUpdate was not updated within timeout, scheduler did not execute job")
}

func TestScheduler_OneTimeJob(t *testing.T) {
	mem.LastUpdate = 0
	ctrl := newTestControllerWithTickMs(t, 250, nil)
	t.Cleanup(func() { ctrl.svc.Exec("exit", multiplex.ExecParams{}) })

	ctrl.ScheduleJob("test_one_time", 0, "Notify", "last_update", multiplex.ExecParams{})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if mem.LastUpdate > 0 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	if mem.LastUpdate == 0 {
		t.Fatal("LastUpdate was not updated, one-time job did not execute")
	}

	firstUpdate := mem.LastUpdate
	mem.LastUpdate = 0

	waitDeadline := time.Now().Add(time.Second)
	for time.Now().Before(waitDeadline) {
		time.Sleep(50 * time.Millisecond)
	}

	if mem.LastUpdate > 0 {
		t.Fatalf("one-time job executed again: first=%d, second=%d", firstUpdate, mem.LastUpdate)
	}
}

func TestScheduler_ScheduleMultipleJobs(t *testing.T) {
	mem.LastUpdate = 0
	ctrl := newTestControllerWithTickMs(t, 250, nil)
	t.Cleanup(func() { ctrl.svc.Exec("exit", multiplex.ExecParams{}) })

	ctrl.ScheduleJob("multi_job_1", 500, "Notify", "last_update", multiplex.ExecParams{})
	ctrl.ScheduleJob("multi_job_2", 500, "Notify", "last_update", multiplex.ExecParams{})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if mem.LastUpdate > 0 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatal("LastUpdate was not updated within timeout, scheduled jobs did not execute")
}

func TestScheduler_UnscheduleStopsExecution(t *testing.T) {
	mem.LastUpdate = 0
	ctrl := newTestControllerWithTickMs(t, 250, nil)
	t.Cleanup(func() { ctrl.svc.Exec("exit", multiplex.ExecParams{}) })

	ctrl.ScheduleJob("test_stop", 500, "Notify", "last_update", multiplex.ExecParams{})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if mem.LastUpdate > 0 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	if mem.LastUpdate == 0 {
		t.Fatal("LastUpdate was not updated before unschedule")
	}

	firstUpdate := mem.LastUpdate
	mem.LastUpdate = 0

	ctrl.UnscheduleJob("test_stop")

	waitDeadline := time.Now().Add(time.Second)
	for time.Now().Before(waitDeadline) {
		time.Sleep(50 * time.Millisecond)
	}

	if mem.LastUpdate > 0 {
		t.Fatalf("job executed after unschedule: first=%d, second=%d", firstUpdate, mem.LastUpdate)
	}
}

func newTestControllerWithTickMs(t *testing.T, tickMs uint64, dbc *db.DbContext) *Controller {
	t.Helper()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: &nullWriter{}, TimeFormat: time.DateTime}).
		With().
		Timestamp().
		Logger()

	cfg := &config.ServiceConfig{
		MaxRpcRrtryCount:         3,
		HistoricalApiWorkerCount: 1,
		StandardApiWorkerCount:   1,
		SchedulerTickMs:          tickMs,
	}
	ctrl := NewController(cfg, nil, dbc, logger)
	go ctrl.Run()
	return ctrl
}
