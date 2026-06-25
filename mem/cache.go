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

package mem

import "github.com/tforce-io/tf-golib/stdx/mathxt"

var (
	BlockHeight = &LatestBlock{
		Block:     0,
		CallTrace: 0,
	}
	LastUpdate = int64(0)
)

type LatestBlock struct {
	Block     uint64
	CallTrace uint64
}

func (s *LatestBlock) Overall() uint64 {
	return mathxt.MinUint64(s.Block, s.CallTrace)
}
