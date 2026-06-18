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

package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

const (
	AddressLength = 20
	Bytes32Length = 32
)

var (
	ZeroAddress = Address{}
	ZeroBytes32 = Bytes32{}
)

// Address represents an Ethereum address.
type Address [AddressLength]byte

// Return new instance of Address from a hex string.
func NewAddressFromString(s string) (Address, error) {
	return decodeAddress(s)
}

// Return Address as a byte slice.
func (a Address) Bytes() []byte {
	b := make([]byte, AddressLength)
	copy(b, a[:])
	return b
}

// Check if the Address is equal to a hex string.
func (a Address) Equal(s string) bool {
	src := strings.ToLower(strings.TrimPrefix(s, "0x"))
	hexSrc := hex.EncodeToString(a[:])
	return src == hexSrc
}

// Return Address as a hex string with 0x prefix.
func (a Address) Hex() string {
	return "0x" + hex.EncodeToString(a[:])
}

// Return Address as a hex string without prefix.
func (a Address) String() string {
	return a.Hex()
}

// Implement json.Marshaler.
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Hex())
}

// Implement json.Unmarshaler.
func (a *Address) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	addr, err := decodeAddress(s)
	if err != nil {
		return err
	}
	*a = addr
	return nil
}

// Bytes represents a variable-length hex-encoded byte sequence.
type Bytes []byte

// Return Bytes as a byte slice.
func (b Bytes) Bytes() []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// Return Bytes as a hex string with 0x prefix.
func (b Bytes) Hex() string {
	if len(b) == 0 {
		return "0x"
	}
	return "0x" + hex.EncodeToString([]byte(b))
}

// Return Bytes as a hex string without prefix.
func (b Bytes) String() string {
	return b.Hex()
}

// Implement json.Marshaler.
func (b *Bytes) MarshalJSON() ([]byte, error) {
	if b == nil || len(*b) == 0 {
		return json.Marshal("0x")
	}
	return json.Marshal(b.Hex())
}

// Implement json.Unmarshaler.
func (b *Bytes) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n, err := decodeHex(s)
	if err != nil {
		return err
	}
	*b = n
	return nil
}

// Bytes32 represents a fixed length of 32 byte sequence.
type Bytes32 [Bytes32Length]byte

// Return new instance of Bytes32 from a hex string.
func NewBytes32FromString(s string) (Bytes32, error) {
	return decodeBytes32(s)
}

// Return Bytes32 as a byte slice.
func (b Bytes32) Bytes() []byte {
	c := make([]byte, Bytes32Length)
	copy(c, b[:])
	return c
}

// Check if the Bytes32 is equal to a hex string.
func (b Bytes32) Equal(s string) bool {
	src := strings.ToLower(strings.TrimPrefix(s, "0x"))
	hexSrc := hex.EncodeToString(b[:])
	return src == hexSrc
}

// Return Bytes32 as a hex string with 0x prefix.
func (b Bytes32) Hex() string {
	return "0x" + hex.EncodeToString(b[:])
}

// Return Bytes32 as a hex string without prefix.
func (b Bytes32) String() string {
	return b.Hex()
}

// Implement json.Marshaler.
func (b Bytes32) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Hex())
}

// Implement json.Unmarshaler.
func (b *Bytes32) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n, err := decodeBytes32(s)
	if err != nil {
		return err
	}
	*b = n
	return nil
}

// Integer represents a hex-encoded integer.
type Integer big.Int

// Return Integer as a hex string with 0x prefix.
func (i *Integer) Hex() string {
	if i == nil || (*big.Int)(i).Sign() == 0 {
		return "0x0"
	}
	return "0x" + (*big.Int)(i).Text(16)
}

// Return Integer as a *big.Int.
func (i *Integer) Int() *big.Int {
	if i == nil {
		return big.NewInt(0)
	}
	result := new(big.Int)
	result.Set((*big.Int)(i))
	return result
}

// Return Integer as a hex string without prefix.
func (i *Integer) String() string {
	return i.Hex()
}

// Return Integer as a uint64.
func (i *Integer) Uint64() uint64 {
	if i == nil {
		return 0
	}
	return (*big.Int)(i).Uint64()
}

// Implement json.Marshaler.
func (i *Integer) MarshalJSON() ([]byte, error) {
	if i == nil || (*big.Int)(i).Sign() == 0 {
		return json.Marshal("0x0")
	}
	return json.Marshal("0x" + (*big.Int)(i).Text(16))
}

// Implement json.Unmarshaler.
func (i *Integer) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := decodeNumericString(s)
	if err != nil {
		return err
	}
	if val == nil {
		*i = Integer{}
		return nil
	}
	*i = Integer(*val)
	return nil
}

// Return new instance of Address from a hex string.
func decodeAddress(s string) (Address, error) {
	if s == "" || s == "0x" {
		return ZeroAddress, nil
	}
	src := strings.TrimPrefix(s, "0x")
	if len(src) != AddressLength*2 {
		return ZeroAddress, fmt.Errorf("decode address %s: invalid length %d, expected %d", s, len(src), AddressLength*2)
	}
	b, err := decodeHex(s)
	if err != nil {
		return ZeroAddress, err
	}
	var addr Address
	copy(addr[AddressLength-len(b):], b)
	return addr, nil
}

// Return new instance of Bytes32 from a hex string.
func decodeBytes32(s string) (Bytes32, error) {
	if s == "" || s == "0x" {
		return ZeroBytes32, nil
	}
	src := strings.TrimPrefix(s, "0x")
	if len(src) != Bytes32Length*2 {
		return ZeroBytes32, fmt.Errorf("decode bytes32 %s: invalid length %d, expected %d", s, len(src), Bytes32Length*2)
	}
	b, err := decodeHex(s)
	if err != nil {
		return ZeroBytes32, err
	}
	var h Bytes32
	copy(h[Bytes32Length-len(b):], b)
	return h, nil
}
