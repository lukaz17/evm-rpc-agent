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

// RootConfig contains all available configurations for the application.
type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	Log        *LogConfig    `koanf:"log"`
	RPC        *RPCConfig    `koanf:"rpc"`
	Server     *ServerConfig `koanf:"server"`
}

// LogConfig contains configurations for logging behavior.
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
	ToFile bool   `koanf:"to_file"`
	LogDir string `koanf:"log_dir"`
}

// RPCConfig contains configurations for EVM RPC connections.
type RPCConfig struct {
	URL     string `koanf:"url"`
	ChainID int64  `koanf:"chain_id"`
}

// ServerConfig contains configurations for the HTTP server.
type ServerConfig struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}
