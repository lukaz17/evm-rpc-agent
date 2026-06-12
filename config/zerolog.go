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

package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init logging for the application.
func InitZerolog(configDir string, useFS bool) (zerolog.Logger, *lumberjack.Logger, error) {
	colorSupported := isatty.IsTerminal(os.Stdout.Fd())

	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: zerolog.ConsoleWriter{
				Out:        os.Stdout,
				NoColor:    !colorSupported,
				TimeFormat: time.DateTime,
			},
		},
		Level: zerolog.TraceLevel,
	}

	logWriter := initLogFile(useFS, configDir)
	if logWriter == nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		return consoleLogger, nil, nil
	}

	fileWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: logWriter,
		},
		Level: zerolog.TraceLevel,
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	return logger, logWriter, nil
}

// Create log file with rotation support.
func initLogFile(useFS bool, workingDir string) *lumberjack.Logger {
	if !useFS {
		return nil
	}

	logDir := workingDir
	if logDir == "" {
		logDir = "."
	}
	logDir = filepath.Join(logDir, "logs")

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil
		}
	}

	logFilePath := filepath.Join(logDir, "evm-rpc-agent.log")

	return &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
		LocalTime:  true,
	}
}
