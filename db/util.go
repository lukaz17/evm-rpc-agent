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

package db

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Convert bson.D to json.RawMessage.
func BsonToJson(doc bson.D) (json.RawMessage, error) {
	data, err := bson.MarshalExtJSON(doc, true, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Convert bson.A to json.RawMessage.
func BsonArrToJson(arr bson.A) (json.RawMessage, error) {
	data, err := bson.MarshalExtJSON(arr, true, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Convert json.RawMessage to bson.D.
func JsonToBson(rawJson json.RawMessage) (bson.D, error) {
	var doc bson.D
	if err := bson.UnmarshalExtJSON(rawJson, true, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Convert json.RawMessage to bson.A.
func JsonToBsonArr(rawJson json.RawMessage) (bson.A, error) {
	var arr bson.A
	if err := bson.UnmarshalExtJSON(rawJson, true, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

func getNumberFromDoc(doc bson.D, key string) (uint64, error) {
	var blockNum uint64
	err := fmt.Errorf("key not found or invalid type: %s", key)
	for _, e := range doc {
		if e.Key == key {
			switch v := e.Value.(type) {
			case int:
			case int8:
			case int16:
			case int32:
			case int64:
			case uint:
			case uint8:
			case uint16:
			case uint32:
			case uint64:
				blockNum = uint64(v)
				err = nil
			case string:
				fmt.Sscanf(v, "0x%x", &blockNum)
				err = nil
			}
			break
		}
	}
	if err != nil {
		return 0, err
	}
	return blockNum, nil
}
