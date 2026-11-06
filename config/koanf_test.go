// Copyright (C) 2025  Nguyen Nhat Tung
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"testing"
)

func TestDefaultConfig_TraceTimeoutSeconds(t *testing.T) {
	got, ok := DefaultConfig().Get("service.trace_timeout_seconds").(uint64)
	if !ok {
		t.Fatalf("service.trace_timeout_seconds is not uint64, got %T", got)
	}

	if got != 120 {
		t.Errorf("service.trace_timeout_seconds = %d, want 120", got)
	}
}
