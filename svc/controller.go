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
	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

// Controller wires up and manage all services in the EVM RPC agent.
type Controller struct {
	cfg    *config.ServiceConfig
	rpc    *rpc.Client
	svc    *multiplex.ServiceController
	logger diag.Logger
}

// Return new Controller instance.
func NewController(cfg *config.ServiceConfig, rpc *rpc.Client, dbc *db.DbContext, logger zerolog.Logger) *Controller {
	diagLogger := config.ZerologAdapter{Logger: logger}
	router := multiplex.NewServiceController(diagLogger)
	scheduler := NewScheduler(cfg.SchedulerTickMs, logger)
	scheduler.SetRouter(router)
	scheduler.SetWorker(1)
	router.Register(scheduler)

	callEthApiSvc := NewCallEthApi(logger)
	callEthApiSvc.SetRouter(router)
	callEthApiSvc.SetWorker(4)
	router.Register(callEthApiSvc)

	getBlockSvc := NewGetBlock(logger)
	getBlockSvc.SetRouter(router)
	getBlockSvc.SetWorker(2)
	router.Register(getBlockSvc)
	notifySvc := NewNotify(logger)
	notifySvc.SetRouter(router)
	notifySvc.SetWorker(1)
	router.Register(notifySvc)

	traceBlockSvc := NewTraceBlock(logger)
	traceBlockSvc.SetRouter(router)
	traceBlockSvc.SetWorker(2)
	router.Register(traceBlockSvc)

	if dbc == nil {
		diagLogger.Warn("Database services are not available.")
	} else {
		writeDatabaseSvc := NewWriteDatabase(dbc, logger)
		writeDatabaseSvc.SetRouter(router)
		writeDatabaseSvc.SetWorker(1)
		router.Register(writeDatabaseSvc)
	}

	if rpc == nil {
		diagLogger.Warn("RPC services are not available.")
	} else {
		exclCallEthRpcSvc := NewCallEthRpc("ExclCallEthRpc", rpc, cfg, logger)
		exclCallEthRpcSvc.SetRouter(router)
		exclCallEthRpcSvc.SetWorker(1)
		router.Register(exclCallEthRpcSvc)

		crawlEthDataSvc := NewCrawlEthData(cfg.CrawlBatchSize, logger)
		crawlEthDataSvc.SetRouter(router)
		crawlEthDataSvc.SetWorker(1)
		router.Register(crawlEthDataSvc)

		histCallEthRpcSvc := NewCallEthRpc("HistCallEthRpc", rpc, cfg, logger)
		histCallEthRpcSvc.SetRouter(router)
		histCallEthRpcSvc.SetWorker(uint64(cfg.HistoricalApiWorkerCount))
		router.Register(histCallEthRpcSvc)

		stdCallEthRpcSvc := NewCallEthRpc("StdCallEthRpc", rpc, cfg, logger)
		stdCallEthRpcSvc.SetRouter(router)
		stdCallEthRpcSvc.SetWorker(uint64(cfg.StandardApiWorkerCount))
		router.Register(stdCallEthRpcSvc)
	}

	return &Controller{
		cfg:    cfg,
		rpc:    rpc,
		svc:    router,
		logger: diagLogger,
	}
}

// Start listening and routing messages.
func (c *Controller) Run() {
	c.svc.Run(true)
}

// Dispatch command to a service.
func (c *Controller) Dispatch(serviceID string, command string, params multiplex.ExecParams) {
	c.svc.Dispatch(serviceID, command, params)
}

// Dispatch command to a service and wait for it to complete. Controller will stop once the command is processed.
func (c *Controller) DispatchOnce(serviceID string, command string, params multiplex.ExecParams) {
	params.ExpectReturn()
	c.svc.Dispatch(serviceID, command, params)
	params.Wait()
	c.svc.Exec("exit", multiplex.ExecParams{})
}

// Schedule a recurring job.
func (c *Controller) ScheduleJob(jobID string, intervalMs int64, serviceID string, command string, params multiplex.ExecParams) {
	p := multiplex.ExecParams{
		"job_id":      jobID,
		"interval_ms": intervalMs,
		"service_id":  serviceID,
		"command":     command,
		"params":      params,
	}
	c.svc.Dispatch("Scheduler", "schedule_job", p)
}

// Remove a recurring job.
func (c *Controller) UnscheduleJob(jobID string) {
	p := multiplex.ExecParams{
		"job_id": jobID,
	}
	c.svc.Dispatch("Scheduler", "unschedule_job", p)
}
