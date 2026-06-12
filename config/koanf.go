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
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

var cfg *RootConfig

// Init configuration for the application.
func InitKoanf(useFS bool) (*RootConfig, error) {
	if cfg != nil {
		return cfg, nil
	}

	isPortable := !useFS || IsPortable()
	configFile := "evm-rpc-agent.yml"

	if isPortable {
		exec, _ := os.Executable()
		configFile = path.Join(path.Dir(exec), "evm-rpc-agent.yml")
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		configFile = path.Join(home, ".config", "evm-rpc-agent", "evm-rpc-agent.yml")
	} else if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		configFile = path.Join(appData, "evm-rpc-agent", "evm-rpc-agent.yml")
	}

	var err error
	cfg, err = buildConfig(useFS, configFile)
	if err != nil {
		return cfg, err
	}

	cfg.ConfigDir = path.Dir(configFile)
	cfg.ConfigFile = configFile
	cfg.IsPortable = isPortable
	return cfg, nil
}

// Check if the application is in portable mode.
func IsPortable() bool {
	exec, _ := os.Executable()
	portableFile := filepath.Join(filepath.Dir(exec), "evm-rpc-agent.portable")
	return fileExists(portableFile)
}

// Build configurations for the application with the following priority:
// Environment variables -> YAML configuration file -> default values.
func buildConfig(useFS bool, f string) (*RootConfig, error) {
	k := defaultConfig()
	if useFS && fileExists(f) {
		k, _ = configFromYaml(k, f)
	}
	k, _ = configFromEnv(k)

	var config RootConfig
	err := k.Unmarshal("", &config)
	return &config, err
}

// Get default configuration values.
func defaultConfig() *koanf.Koanf {
	k := koanf.New(".")

	k.Load(
		structs.Provider(RootConfig{
			Log: &LogConfig{
				Level:  "info",
				Format: "console",
				ToFile: false,
				LogDir: "",
			},
			RPC: &RPCConfig{
				URL:     "http://localhost:8545",
				ChainID: 1,
			},
			Server: &ServerConfig{
				Host: "127.0.0.1",
				Port: 8080,
			},
		}, "koanf"),
		nil,
	)

	return k
}

// Get configuration values from environment variables.
func configFromEnv(k *koanf.Koanf) (*koanf.Koanf, error) {
	err := k.Load(env.Provider("LKZ_EVM_RPC_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(
				strings.TrimPrefix(s, "LKZ_EVM_RPC_")), "_", ".", -1)
	}), nil)
	if err != nil {
		return k, err
	}
	return k, nil
}

// Get configuration values from YAML file.
func configFromYaml(k *koanf.Koanf, f string) (*koanf.Koanf, error) {
	err := k.Load(file.Provider(f), yaml.Parser())
	if err != nil {
		return k, err
	}
	return k, nil
}

// Return whether a file existed in file system.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
