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
	"math/rand"
	"time"
)

// NetworkOptions define the parameters for network calls with retry logic.
type NetworkOptions struct {
	MaxRetries  int
	MaxRetryGap uint64
}

// Sleep for a random duration between 20ms and MaxRetryGap.
func (o *NetworkOptions) WaitRetryGap() {
	minRetryGap := uint64(20000000)
	if o.MaxRetryGap < minRetryGap {
		time.Sleep(time.Duration(minRetryGap))
	}
	duration := uint64(rand.Int63n(int64(o.MaxRetryGap-minRetryGap/2+1))) + minRetryGap/2
	time.Sleep(time.Duration(duration))
}
