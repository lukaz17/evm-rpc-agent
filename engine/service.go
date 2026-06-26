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

package engine

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/db"
	"github.com/lukaz17/evm-rpc-agent/mem"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/lukaz17/evm-rpc-agent/svc"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/multiplex"
)

// ServiceModule handles the EVM data crawler daemon.
type ServiceModule struct {
	logger zerolog.Logger
}

// Return new ServiceModule instance.
func NewServiceModule(c *Controller) *ServiceModule {
	return &ServiceModule{
		logger: c.ModuleLogger("service"),
	}
}

// Start the EVM data crawler daemon and handle graceful shutdown.
func (m *ServiceModule) Run(cfg *config.RootConfig) error {
	m.logger.Info().Msg("Starting EVM RPC Agent service...")

	rpcClient := rpc.NewClient(cfg.RPC.URL)
	dbc := db.NewDbContext(cfg.Database)
	if err := dbc.Connect(); err != nil {
		return err
	}
	defer func() {
		m.logger.Info().Msg("Disconnecting from MongoDB...")
		_ = dbc.Disconnect()
		m.logger.Info().Msg("MongoDB disconnected.")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	blockCount, err := dbc.CountBlocks(ctx)
	if err != nil {
		m.logger.Warn().Err(err).Msg("Failed to count blocks collection.")
	} else {
		mem.CurrentHeight.Block = uint64(blockCount)
		m.logger.Info().Uint64("count", mem.CurrentHeight.Block).Msg("Pre-populated block height from database.")
	}
	traceCount, err := dbc.CountCallTraces(ctx)
	if err != nil {
		m.logger.Warn().Err(err).Msg("Failed to count call traces collection.")
	} else {
		mem.CurrentHeight.CallTrace = uint64(traceCount)
		m.logger.Info().Uint64("count", mem.CurrentHeight.CallTrace).Msg("Pre-populated call trace height from database.")
	}
	cancel()

	svcCfg := &config.ServiceConfig{
		CrawlIntervalMs:  cfg.Service.CrawlIntervalMs,
		CrawlBatchSize:   cfg.Service.CrawlBatchSize,
		MaxRpcRrtryCount: cfg.Service.MaxRpcRrtryCount,
		SchedulerTickMs:  cfg.Service.SchedulerTickMs,

		HistoricalApiWorkerCount: cfg.Service.HistoricalApiWorkerCount,
		StandardApiWorkerCount:   cfg.Service.StandardApiWorkerCount,
	}

	controller := svc.NewController(svcCfg, rpcClient, dbc, m.logger)

	intervalMs := svcCfg.CrawlIntervalMs
	if intervalMs <= 0 {
		intervalMs = 1000
		m.logger.Warn().Int64("default", intervalMs).Msg("CrawlIntervalMs is zero or negative, using default.")
	}
	batchSize := svcCfg.CrawlBatchSize
	if batchSize <= 0 {
		batchSize = 10
		m.logger.Warn().Int("default", batchSize).Msg("CrawlBatchSize is zero or negative, using default.")
	}

	crawlMsg := multiplex.ExecParams{
		"batch_size": batchSize,
	}
	controller.ScheduleJob("crawl_eth_data", intervalMs, "CrawlEthData", "crawl", crawlMsg)
	controller.ScheduleJob("height_status", 10000, "Notify", "height_status", multiplex.ExecParams{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		m.logger.Info().Msg("Shutdown signal received. Stopping service...")
		controller.UnscheduleJob("crawl_eth_data")
		controller.UnscheduleJob("height_status")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.DispatchOnce("Scheduler", "exit", multiplex.ExecParams{})
		go func() {
			<-ctx.Done()
			m.logger.Error().Msg("Force shutdown due to timeout.")
		}()
	}()
	controller.Run()

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *ServiceModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred.")
	}
}

// Define Cobra Command for Service module.
func ServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Start EVM data crawler daemon.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			m := NewServiceModule(c)

			if err := m.Run(c.Root); err != nil {
				m.logError(err)
			}
		},
	}
	return cmd
}

// ServiceFlags contains all flags used by Service module.
type ServiceFlags struct {
}

// Extract all flags from a Cobra Command.
func (m *ServiceModule) ParseServiceFlags(cmd *cobra.Command, args []string) *ServiceFlags {

	f := &ServiceFlags{}
	return f
}
