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

package xlib

import "github.com/rs/zerolog"

// Bytes implements zerolog.LogObjectMarshaler for []byte.
type Bytes []byte

func (slice Bytes) MarshalZerologArray(arr *zerolog.Array) {
	arr.Bytes(slice)
}

// IntSlice implements zerolog.LogArrayMarshaler for []int.
type IntSlice []int

func (slice IntSlice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Int(i)
	}
}

// Int64Slice implements zerolog.LogArrayMarshaler for []int64.
type Int64Slice []int64

func (slice Int64Slice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Int64(i)
	}
}

// StringSlice implements zerolog.LogArrayMarshaler for []string.
type StringSlice []string

func (slice StringSlice) MarshalZerologArray(arr *zerolog.Array) {
	for _, s := range slice {
		arr.Str(s)
	}
}
