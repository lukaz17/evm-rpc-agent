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

package core

import (
	"bytes"
	"math/big"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewAddressFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Address
		wantErr  bool
	}{
		{"empty", "", ZeroAddress, false},
		{"0x_prefix_only", "0x", ZeroAddress, false},
		{"valid_with_0x", "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
		{"valid_without_0x", "de0b295669a9fd93d5f28d9ec85e40f4cb697bae", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
		{"uppercase", "0xDE0B295669A9FD93D5F28D9EC85E40F4CB697BAE", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
		{"zero_address", "0x0000000000000000000000000000000000000000", ZeroAddress, false},
		{"too_short", "0xde0b295669a9fd93d5f28d9ec85e40f4cb697ba", ZeroAddress, true},
		{"too_long", "0xde0b295669a9fd93d5f28d9ec85e40f4cb697baee", ZeroAddress, true},
		{"invalid_hex", "0xGG0B295669A9FD93D5F28D9EC85E40F4CB697BAE", ZeroAddress, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewAddressFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), result.Hex())
			}
		})
	}
}

func TestAddress_Getters(t *testing.T) {
	tests := []struct {
		name      string
		addr      Address
		wantHex   string
		wantByte  byte
		byteIndex int
	}{
		{"zero", ZeroAddress, "0x0000000000000000000000000000000000000000", 0x00, 19},
		{"zero_leading_address", Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56}, "0x000000000000000000000000ff00abcdef123456", 0x56, 19},
		{"full_address", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae", 0xae, 19},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.addr.Hex(); got != tt.wantHex {
				t.Errorf("Address.Hex() = %s, want %s", got, tt.wantHex)
			}
			if got := tt.addr.String(); got != tt.wantHex {
				t.Errorf("Address.String() = %s, want %s", got, tt.wantHex)
			}
			b := tt.addr.Bytes()
			if len(b) != AddressLength {
				t.Errorf("expected Bytes length %d, got %d", AddressLength, len(b))
			}
			if b[tt.byteIndex] != tt.wantByte {
				t.Errorf("expected byte[%d]=0x%02x, got 0x%02x", tt.byteIndex, tt.wantByte, b[tt.byteIndex])
			}
		})
	}
}

func TestAddress_Equal(t *testing.T) {
	addr := Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}

	tests := []struct {
		name     string
		addr     Address
		input    string
		expected bool
	}{
		{"exact match with prefix", addr, "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae", true},
		{"exact match no prefix", addr, "de0b295669a9fd93d5f28d9ec85e40f4cb697bae", true},
		{"uppercase match", addr, "0xDE0B295669A9FD93D5F28D9EC85E40F4CB697BAE", true},
		{"mixed case match", addr, "0xDe0b295669a9Fd93D5f28d9Ec85e40F4Cb697baE", true},
		{"mismatch address", addr, "0x0000000000000000000000000000000000000001", false},
		{"wrong length", addr, "0xde0b295669a9fd93d5f28d9ec85e40f4cb697b", false},
		{"zero address match", ZeroAddress, "0x0000000000000000000000000000000000000000", true},
		{"zero address mismatch", ZeroAddress, "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.addr.Equal(tt.input); got != tt.expected {
				t.Errorf("Address.Equal(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAddress_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		addr     Address
		expected string
	}{
		{"zero", ZeroAddress, `"0x0000000000000000000000000000000000000000"`},
		{"full_address", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, `"0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.addr.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestAddress_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Address
		wantErr  bool
	}{
		{"valid address", `"0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"`, Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
		{"zero address", `"0x0000000000000000000000000000000000000000"`, ZeroAddress, false},
		{"invalid length", `"0xde0b295669a9fd93d5f28d9ec85e40f4cb697b"`, ZeroAddress, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addr Address
			err := addr.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if addr != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), addr.Hex())
			}
		})
	}
}

func TestAddress_MarshalBSONValue(t *testing.T) {
	tests := []struct {
		name     string
		addr     Address
		wantType byte
	}{
		{"zero", ZeroAddress, byte(bson.TypeString)},
		{"full_address", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, byte(bson.TypeString)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.addr.MarshalBSONValue()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if typ != tt.wantType {
				t.Errorf("expected BSON type %v, got %v", tt.wantType, typ)
			}
			if len(data) == 0 {
				t.Error("expected non-empty BSON data")
			}
		})
	}
}

func TestAddress_UnmarshalBSONValue(t *testing.T) {
	tests := []struct {
		name    string
		input   Address
		wantErr bool
	}{
		{"zero", ZeroAddress, false},
		{"full_address", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.input.MarshalBSONValue()
			if err != nil {
				t.Fatalf("MarshalBSONValue error: %v", err)
			}
			var got Address
			if err := got.UnmarshalBSONValue(typ, data); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("UnmarshalBSONValue error: %v", err)
			}
			if got != tt.input {
				t.Errorf("roundtrip mismatch: got %s, want %s", got.Hex(), tt.input.Hex())
			}
		})
	}
}

func TestAddress_BSONRoundtrip(t *testing.T) {
	type doc struct {
		Addr Address `bson:"addr"`
	}
	original := doc{Addr: Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}}

	data, err := bson.Marshal(original)
	if err != nil {
		t.Fatalf("bson.Marshal error: %v", err)
	}

	var result doc
	if err := bson.Unmarshal(data, &result); err != nil {
		t.Fatalf("bson.Unmarshal error: %v", err)
	}

	if result.Addr != original.Addr {
		t.Errorf("roundtrip mismatch: got %s, want %s", result.Addr.Hex(), original.Addr.Hex())
	}
}

func TestAddress_ZeroAddress(t *testing.T) {
	addr := ZeroAddress
	if addr.Hex() != "0x0000000000000000000000000000000000000000" {
		t.Errorf("expected zero address hex, got %s", addr.Hex())
	}
}

func TestNewBytesFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Bytes
		wantErr  bool
	}{
		{"empty", "", nil, false},
		{"0x_prefix_only", "0x", nil, false},
		{"with_0x", "0x12ab", Bytes{0x12, 0xab}, false},
		{"without_0x", "12ab", Bytes{0x12, 0xab}, false},
		{"odd_length", "0x123", Bytes{0x01, 0x23}, false},
		{"invalid_hex", "0xGG", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewBytesFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBytes_Getters(t *testing.T) {
	h := Bytes{0xff, 0x00, 0xab}

	b := h.Bytes()
	if len(b) != 3 || b[0] != 0xff {
		t.Errorf("Bytes mismatch: %v", b)
	}

	if h.Hex() != "0xff00ab" {
		t.Errorf("expected 0xff00ab, got %s", h.Hex())
	}

	if h.String() != "0xff00ab" {
		t.Errorf("expected String to match Hex: %s", h.String())
	}
}

func TestBytes_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		hex      Bytes
		expected string
	}{
		{"empty", nil, `"0x"`},
		{"zero", Bytes{}, `"0x"`},
		{"value", Bytes{0x12, 0xab}, `"0x12ab"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.hex.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestBytes_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Bytes
		wantErr  bool
	}{
		{"empty", `"0x"`, nil, false},
		{"value", `"0x12ab"`, Bytes{0x12, 0xab}, false},
		{"no prefix", `"12ab"`, Bytes{0x12, 0xab}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h Bytes
			err := h.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if string(h) != string(tt.expected) && !(h == nil && tt.expected == nil) {
				t.Errorf("expected %v, got %v", tt.expected, h)
			}
		})
	}
}

func TestBytes_MarshalBSONValue(t *testing.T) {
	tests := []struct {
		name     string
		b        Bytes
		wantType byte
		wantNull bool
	}{
		{"nil", nil, byte(bson.TypeNull), true},
		{"empty", Bytes{}, byte(bson.TypeNull), true},
		{"value", Bytes{0x12, 0xab}, byte(bson.TypeString), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.b.MarshalBSONValue()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNull {
				if typ != byte(bson.TypeNull) {
					t.Errorf("expected TypeNull, got %v", typ)
				}
				if data != nil {
					t.Errorf("expected nil data for null, got %v", data)
				}
			} else {
				if typ != tt.wantType {
					t.Errorf("expected BSON type %v, got %v", tt.wantType, typ)
				}
				if len(data) == 0 {
					t.Error("expected non-empty BSON data")
				}
			}
		})
	}
}

func TestBytes_UnmarshalBSONValue(t *testing.T) {
	tests := []struct {
		name    string
		input   Bytes
		wantNil bool
	}{
		{"nil (null)", nil, true},
		{"empty (null)", Bytes{}, true},
		{"value", Bytes{0x12, 0xab}, false},
		{"longer value", Bytes{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.input.MarshalBSONValue()
			if err != nil {
				t.Fatalf("MarshalBSONValue error: %v", err)
			}
			var got Bytes
			if err := got.UnmarshalBSONValue(typ, data); err != nil {
				t.Fatalf("UnmarshalBSONValue error: %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil Bytes, got %v", got)
				}
			} else {
				if !bytes.Equal(got, tt.input) {
					t.Errorf("roundtrip mismatch: got %s, want %s", got.Hex(), tt.input.Hex())
				}
			}
		})
	}
}

func TestBytes_BSONRoundtrip(t *testing.T) {
	type doc struct {
		Data Bytes `bson:"data"`
	}
	original := doc{Data: Bytes{0xde, 0xad, 0xbe, 0xef}}

	data, err := bson.Marshal(original)
	if err != nil {
		t.Fatalf("bson.Marshal error: %v", err)
	}

	var result doc
	if err := bson.Unmarshal(data, &result); err != nil {
		t.Fatalf("bson.Unmarshal error: %v", err)
	}

	if !bytes.Equal(result.Data, original.Data) {
		t.Errorf("roundtrip mismatch: got %s, want %s", result.Data.Hex(), original.Data.Hex())
	}
}

func TestNewBytes32FromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Bytes32
		wantErr  bool
	}{
		{"valid value", "0x0001a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e", Bytes32{0x00, 0x01, 0xa6, 0xd4, 0x0b, 0xf4, 0x20, 0x40, 0x4a, 0x01, 0x17, 0x33, 0xcf, 0xb7, 0xb1, 0x90, 0xd6, 0x2c, 0x65, 0xbf, 0x0b, 0xcd, 0xa3, 0x2b, 0x57, 0xb2, 0x77, 0xd9, 0xad, 0x9f, 0x14, 0x6e}, false},
		{"zero value", "0x0000000000000000000000000000000000000000000000000000000000000000", ZeroBytes32, false},
		{"invalid length", "0xdeadbeef", ZeroBytes32, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewBytes32FromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), result.Hex())
			}
		})
	}
}

func TestBytes32_Getters(t *testing.T) {
	tests := []struct {
		name      string
		b32       Bytes32
		wantHex   string
		wantByte  byte
		byteIndex int
	}{
		{"zero", ZeroBytes32, "0x0000000000000000000000000000000000000000000000000000000000000000", 0x00, 31},
		{"leading_zeros", Bytes32{0x00, 0x01, 0xa6, 0xd4, 0x0b, 0xf4, 0x20, 0x40, 0x4a, 0x01, 0x17, 0x33, 0xcf, 0xb7, 0xb1, 0x90, 0xd6, 0x2c, 0x65, 0xbf, 0x0b, 0xcd, 0xa3, 0x2b, 0x57, 0xb2, 0x77, 0xd9, 0xad, 0x9f, 0x14, 0x6e}, "0x0001a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e", 0x6e, 31},
		{"full_value", Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}, "0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", 0x06, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b32.Hex(); got != tt.wantHex {
				t.Errorf("Bytes32.Hex() = %s, want %s", got, tt.wantHex)
			}
			if got := tt.b32.String(); got != tt.wantHex {
				t.Errorf("Bytes32.String() = %s, want %s", got, tt.wantHex)
			}
			b := tt.b32.Bytes()
			if len(b) != Bytes32Length {
				t.Errorf("expected Bytes length %d, got %d", Bytes32Length, len(b))
			}
			if b[tt.byteIndex] != tt.wantByte {
				t.Errorf("expected byte[%d]=0x%02x, got 0x%02x", tt.byteIndex, tt.wantByte, b[tt.byteIndex])
			}
		})
	}
}

func TestBytes32_Equal(t *testing.T) {
	b32 := Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}

	tests := []struct {
		name     string
		b32      Bytes32
		input    string
		expected bool
	}{
		{"exact match with prefix", b32, "0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", true},
		{"exact match no prefix", b32, "df88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", true},
		{"uppercase match", b32, "0xDF88D4E552ED2E9AFFC3EC61DECE4C69E16D884F6DD5D063E539796055FC3B06", true},
		{"mixed case match", b32, "0xDf88d4e552Ed2e9aFFc3Ec61DecE4C69E16D884F6dD5D063e539796055fc3B06", true},
		{"mismatch value", b32, "0x0000000000000000000000000000000000000000000000000000000000000001", false},
		{"wrong length", b32, "0xdf88d4e552ed2e9affc3ec61dece4c69", false},
		{"zero bytes32 match", ZeroBytes32, "0x0000000000000000000000000000000000000000000000000000000000000000", true},
		{"zero bytes32 mismatch", ZeroBytes32, "0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b32.Equal(tt.input); got != tt.expected {
				t.Errorf("Bytes32.Equal(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBytes32_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		b32      Bytes32
		expected string
	}{
		{"zero", ZeroBytes32, `"0x0000000000000000000000000000000000000000000000000000000000000000"`},
		{"full_value", Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}, `"0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.b32.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestBytes32_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Bytes32
		wantErr  bool
	}{
		{"valid value", `"0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06"`, Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}, false},
		{"zero value", `"0x0000000000000000000000000000000000000000000000000000000000000000"`, ZeroBytes32, false},
		{"invalid length", `"0xdf88d4e552ed2e9affc3ec61dece4c69"`, ZeroBytes32, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b32 Bytes32
			err := b32.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if b32 != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), b32.Hex())
			}
		})
	}
}

func TestBytes32_MarshalBSONValue(t *testing.T) {
	tests := []struct {
		name     string
		b32      Bytes32
		wantType byte
	}{
		{"zero", ZeroBytes32, byte(bson.TypeString)},
		{"full_value", Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}, byte(bson.TypeString)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.b32.MarshalBSONValue()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if typ != tt.wantType {
				t.Errorf("expected BSON type %v, got %v", tt.wantType, typ)
			}
			if len(data) == 0 {
				t.Error("expected non-empty BSON data")
			}
		})
	}
}

func TestBytes32_UnmarshalBSONValue(t *testing.T) {
	tests := []struct {
		name  string
		input Bytes32
	}{
		{"zero", ZeroBytes32},
		{"full_value", Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.input.MarshalBSONValue()
			if err != nil {
				t.Fatalf("MarshalBSONValue error: %v", err)
			}
			var got Bytes32
			if err := got.UnmarshalBSONValue(typ, data); err != nil {
				t.Fatalf("UnmarshalBSONValue error: %v", err)
			}
			if got != tt.input {
				t.Errorf("roundtrip mismatch: got %s, want %s", got.Hex(), tt.input.Hex())
			}
		})
	}
}

func TestBytes32_BSONRoundtrip(t *testing.T) {
	type doc struct {
		Hash Bytes32 `bson:"hash"`
	}
	original := doc{Hash: Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}}

	data, err := bson.Marshal(original)
	if err != nil {
		t.Fatalf("bson.Marshal error: %v", err)
	}

	var result doc
	if err := bson.Unmarshal(data, &result); err != nil {
		t.Fatalf("bson.Unmarshal error: %v", err)
	}

	if result.Hash != original.Hash {
		t.Errorf("roundtrip mismatch: got %s, want %s", result.Hash.Hex(), original.Hash.Hex())
	}
}

func TestBytes32_ZeroBytes32(t *testing.T) {
	b32 := ZeroBytes32
	if b32.Hex() != "0x0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("expected zero bytes32 hex, got %s", b32.Hex())
	}
}

func TestNewIntegerFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Integer
		wantErr  bool
	}{
		{"empty", "", Integer{}, false},
		{"0x_prefix_only", "0x", Integer{}, false},
		{"hex_0x12ab", "0x12ab", Integer(*big.NewInt(0x12ab)), false},
		{"decimal_43981", "43981", Integer(*big.NewInt(0xabcd)), false},
		{"large_hex", "0xffffffffffffffff", Integer(*new(big.Int).SetUint64(0xffffffffffffffff)), false},
		{"invalid", "0xGG", Integer{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewIntegerFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if result.Int().Cmp(tt.expected.Int()) != 0 {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), result.Hex())
			}
		})
	}
}

func TestInteger_Getters(t *testing.T) {
	tests := []struct {
		name       string
		value      Integer
		wantBigInt *big.Int
		wantHex    string
		wantUint64 uint64
	}{
		{"zero", Integer{}, big.NewInt(0), "0x0", uint64(0)},
		{"value_ff00ab", Integer(*big.NewInt(0xff00ab)), big.NewInt(0xff00ab), "0xff00ab", uint64(0xff00ab)},
		{"max_uint64", Integer(*new(big.Int).SetUint64(0xffffffffffffffff)), new(big.Int).SetUint64(0xffffffffffffffff), "0xffffffffffffffff", uint64(0xffffffffffffffff)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Hex(); got != tt.wantHex {
				t.Errorf("Integer.Hex() = %s, want %s", got, tt.wantHex)
			}
			if got := tt.value.String(); got != tt.wantHex {
				t.Errorf("Integer.String() = %s, want %s", got, tt.wantHex)
			}
			if got := tt.value.Uint64(); got != tt.wantUint64 {
				t.Errorf("Integer.Uint64() = 0x%x, want 0x%x", got, tt.wantUint64)
			}
			if got := tt.value.Int(); got.Cmp(tt.wantBigInt) != 0 {
				t.Errorf("Integer.Int() = %s, want %s", got.String(), tt.wantBigInt.String())
			}
		})
	}
}

func TestInteger_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		hex      Integer
		expected string
	}{
		{"zero", Integer{}, `"0x0"`},
		{"value_12ab", Integer(*big.NewInt(0x12ab)), `"0x12ab"`},
		{"large_value", Integer(*new(big.Int).SetUint64(0xffffffffffffffff)), `"0xffffffffffffffff"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.hex.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestInteger_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Integer
		wantErr  bool
	}{
		{"zero", `"0x0"`, Integer{}, false},
		{"empty_0x", `"0x"`, Integer{}, false},
		{"value_12ab", `"0x12ab"`, Integer(*big.NewInt(0x12ab)), false},
		{"no prefix (decimal)", `"43981"`, Integer(*big.NewInt(0xabcd)), false},
		{"large_value", `"0xffffffffffffffff"`, Integer(*new(big.Int).SetUint64(0xffffffffffffffff)), false},
		{"numeric_zero", `0`, Integer{}, false},
		{"numeric_positive", `12345`, Integer(*big.NewInt(12345)), false},
		{"numeric_large", `18446744073709551615`, Integer(*new(big.Int).SetUint64(0xffffffffffffffff)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h Integer
			err := h.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if h.Int().Cmp(tt.expected.Int()) != 0 {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), h.Hex())
			}
		})
	}
}

func TestInteger_MarshalBSONValue(t *testing.T) {
	tests := []struct {
		name     string
		val      Integer
		wantType byte
	}{
		{"zero", Integer{}, byte(bson.TypeInt32)},
		{"small_int32", Integer(*big.NewInt(255)), byte(bson.TypeInt64)},
		{"max_uint64", Integer(*new(big.Int).SetUint64(0xffffffffffffffff)), byte(bson.TypeInt64)},
		{"over_uint64", Integer(*new(big.Int).Lsh(big.NewInt(1), 65)), byte(bson.TypeString)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.val.MarshalBSONValue()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if typ != tt.wantType {
				t.Errorf("expected BSON type %v, got %v", tt.wantType, typ)
			}
			_ = data
		})
	}
}

func TestInteger_UnmarshalBSONValue(t *testing.T) {
	tests := []struct {
		name  string
		input Integer
	}{
		{"zero", Integer{}},
		{"small_value", Integer(*big.NewInt(0x12ab))},
		{"max_uint64", Integer(*new(big.Int).SetUint64(0xffffffffffffffff))},
		{"over_uint64", Integer(*new(big.Int).Lsh(big.NewInt(1), 65))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, data, err := tt.input.MarshalBSONValue()
			if err != nil {
				t.Fatalf("MarshalBSONValue error: %v", err)
			}
			var got Integer
			if err := got.UnmarshalBSONValue(typ, data); err != nil {
				t.Fatalf("UnmarshalBSONValue error: %v", err)
			}
			if got.Int().Cmp(tt.input.Int()) != 0 {
				t.Errorf("roundtrip mismatch: got %s, want %s", got.Hex(), tt.input.Hex())
			}
		})
	}
}

func TestInteger_BSONRoundtrip(t *testing.T) {
	type doc struct {
		Value Integer `bson:"value"`
	}

	cases := []Integer{
		{},
		Integer(*big.NewInt(0xdeadbeef)),
		Integer(*new(big.Int).SetUint64(0xffffffffffffffff)),
		Integer(*new(big.Int).Lsh(big.NewInt(1), 65)),
	}

	for _, c := range cases {
		original := doc{Value: c}
		data, err := bson.Marshal(original)
		if err != nil {
			t.Fatalf("bson.Marshal error for %s: %v", c.Hex(), err)
		}

		var result doc
		if err := bson.Unmarshal(data, &result); err != nil {
			t.Fatalf("bson.Unmarshal error for %s: %v", c.Hex(), err)
		}

		if result.Value.Int().Cmp(original.Value.Int()) != 0 {
			t.Errorf("roundtrip mismatch: got %s, want %s", result.Value.Hex(), original.Value.Hex())
		}
	}
}

func TestInteger_UnmarshalBSONValue_InvalidType(t *testing.T) {
	var i Integer
	err := i.UnmarshalBSONValue(byte(bson.TypeBoolean), []byte{0x01})
	if err == nil {
		t.Error("expected error for unsupported BSON type, got nil")
	}
}

func TestDecodeAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Address
		wantErr  bool
	}{
		{"valid address", "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
		{"no prefix", "de0b295669a9fd93d5f28d9ec85e40f4cb697bae", Address{0xde, 0x0b, 0x29, 0x56, 0x69, 0xa9, 0xfd, 0x93, 0xd5, 0xf2, 0x8d, 0x9e, 0xc8, 0x5e, 0x40, 0xf4, 0xcb, 0x69, 0x7b, 0xae}, false},
		{"zero address", "0x0000000000000000000000000000000000000000", ZeroAddress, false},
		{"empty string", "", ZeroAddress, false},
		{"0x only", "0x", ZeroAddress, false},
		{"too short", "0xde0b295669a9fd93d5f28d9ec85e40f4cb697b", ZeroAddress, true},
		{"too long", "0xde0b295669a9fd93d5f28d9ec85e40f4cb697baeff", ZeroAddress, true},
		{"invalid hex", "0xgg0b295669a9fd93d5f28d9ec85e40f4cb697bae", ZeroAddress, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), result.Hex())
			}
		})
	}
}

func TestDecodeBytes32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Bytes32
		wantErr  bool
	}{
		{"valid value", "0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}, false},
		{"no prefix", "df88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", Bytes32{0xdf, 0x88, 0xd4, 0xe5, 0x52, 0xed, 0x2e, 0x9a, 0xff, 0xc3, 0xec, 0x61, 0xde, 0xce, 0x4c, 0x69, 0xe1, 0x6d, 0x88, 0x4f, 0x6d, 0xd5, 0xd0, 0x63, 0xe5, 0x39, 0x79, 0x60, 0x55, 0xfc, 0x3b, 0x06}, false},
		{"zero value", "0x0000000000000000000000000000000000000000000000000000000000000000", ZeroBytes32, false},
		{"empty string", "", ZeroBytes32, false},
		{"0x only", "0x", ZeroBytes32, false},
		{"too short", "0xdf88d4e552ed2e9affc3ec61dece4c69", ZeroBytes32, true},
		{"too long", "0xdf88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06ff", ZeroBytes32, true},
		{"invalid hex", "0xgg88d4e552ed2e9affc3ec61dece4c69e16d884f6dd5d063e539796055fc3b06", ZeroBytes32, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeBytes32(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: err=%v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.Hex(), result.Hex())
			}
		})
	}
}
