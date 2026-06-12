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
	"os"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Controller is the entrypoint for working with application configurations and loggings.
type Controller struct {
	Root      *config.RootConfig
	Logger    zerolog.Logger
	logWriter *lumberjack.Logger
}

// Entrypoint for creating new instance of Controller.
// useFS will instruct this function to read configurations and create log file.
func NewController(useFS bool) *Controller {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	cfg, err := config.InitKoanf(useFS)
	if err != nil {
		logger.Err(err).Msg("error initializing config")
	}

	var logWriter *lumberjack.Logger
	logger2, logWriter, err2 := config.InitZerolog(cfg.ConfigDir, useFS)
	if err2 != nil {
		logger.Err(err2).Msg("error initializing log file")
	} else {
		logger = logger2
	}

	return &Controller{
		Root:      cfg,
		Logger:    logger,
		logWriter: logWriter,
	}
}

// Execute additional clean up when terminate the app.
func (c *Controller) Close() {
	if c.logWriter != nil {
		c.logWriter.Close()
		c.logWriter = nil
	}
}

// Get a ZeroLog logger instance for command handler from root instance.
func (c *Controller) CommandLogger(module, command string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Str("command", command).Logger()
}

// Get a ZeroLog logger instance for module from root instance.
func (c *Controller) ModuleLogger(module string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Logger()
}
