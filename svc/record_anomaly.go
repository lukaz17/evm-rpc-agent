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

package svc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lukaz17/evm-rpc-agent/config"
	"github.com/lukaz17/evm-rpc-agent/rpc"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/multiplex"
)

// Names of the crawl data anomaly kinds, each recorded in its own file.
const (
	AnomalyTypeEmptyAttestor = "empty_attestor"
	AnomalyTypeTraceTxError  = "trace_tx_error"
)

// anomalyTableHeader is written once at the top of each anomaly file.
const anomalyTableHeader = "| time | block | tx_index | tx_hash | detail |\n|------|-------|----------|---------|--------|\n"

// anomalyTypePattern constrains anomaly type names, which become file names.
var anomalyTypePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// RecordAnomalyParams contains parameters for all commands of RecordAnomaly.
// TxIndex is -1 when the anomaly is not tied to one transaction.
type RecordAnomalyParams struct {
	Type        string
	BlockNumber string
	TxIndex     int
	TxHash      string
	Detail      string
}

// Extract parameters from a service message and return new RecordAnomalyParams.
func NewRecordAnomalyParams(msg *multiplex.ServiceMessage) *RecordAnomalyParams {
	return &RecordAnomalyParams{
		Type:        msg.GetParam("type", "").(string),
		BlockNumber: msg.GetParam("block_number", "").(string),
		TxIndex:     msg.GetParam("tx_index", int(-1)).(int),
		TxHash:      msg.GetParam("tx_hash", "").(string),
		Detail:      msg.GetParam("detail", "").(string),
	}
}

// RecordAnomaly logs crawl data anomalies and appends them to per-type files in the configured directory.
type RecordAnomaly struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	dir string
}

// Return new RecordAnomaly instance. Empty anomalyDir keeps anomalies in the log only.
func NewRecordAnomaly(anomalyDir string, logger zerolog.Logger) *RecordAnomaly {
	svc := &RecordAnomaly{
		dir: anomalyDir,
	}
	svc.i = svc.InitServiceCore("RecordAnomaly", config.ZerologAdapter{Logger: logger}, svc.coreProcessHook)
	return svc
}

// Implement multiplex.ServiceCoreInternal.
func (s *RecordAnomaly) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "record":
		p := NewRecordAnomalyParams(msg)
		s.record(workerID, p)
		msg.Return(true)
	default:
		s.i.Logger.Warnf("%s#%02d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *RecordAnomaly) record(workerID uint64, p *RecordAnomalyParams) {
	s.i.Logger.Warnf("%s#%02d: Anomaly: type=%s block=%s tx_index=%d tx_hash=%s detail=%s.",
		s.i.ServiceID, workerID, p.Type, p.BlockNumber, p.TxIndex, p.TxHash, p.Detail)
	if err := appendAnomaly(s.dir, time.Now().UTC(), p); err != nil {
		s.i.Logger.Warnf("%s#%02d: Anomaly file write failed. Error: %v.", s.i.ServiceID, workerID, err)
	}
}

// Escape a value so it cannot break out of its markdown table cell.
func anomalyCell(value string) string {
	value = strings.ReplaceAll(value, "|", `\|`)
	value = strings.ReplaceAll(value, "\r", " ")
	return strings.ReplaceAll(value, "\n", " ")
}

// Append one anomaly to its per-type markdown file in dir.
// Create the file with a table header on first write.
func appendAnomaly(dir string, at time.Time, p *RecordAnomalyParams) error {
	if dir == "" {
		return nil
	}
	if !anomalyTypePattern.MatchString(p.Type) {
		return fmt.Errorf("invalid anomaly type %q", p.Type)
	}
	if _, parseErr := strconv.ParseUint(p.BlockNumber, 10, 64); parseErr != nil {
		return fmt.Errorf("invalid block number %q", p.BlockNumber)
	}
	if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
		return fmt.Errorf("create anomaly dir %s: %w", dir, mkdirErr)
	}

	filePath := filepath.Join(dir, p.Type+".md")
	needsHeader := true
	if stat, statErr := os.Stat(filePath); statErr == nil && stat.Size() > 0 {
		needsHeader = false
	}
	file, openErr := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		return fmt.Errorf("open anomaly file %s: %w", filePath, openErr)
	}
	defer file.Close()
	if needsHeader {
		if _, writeErr := file.WriteString(anomalyTableHeader); writeErr != nil {
			return fmt.Errorf("write anomaly table header: %w", writeErr)
		}
	}
	row := fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
		at.Format(time.RFC3339),
		p.BlockNumber,
		txIndexCell(p.TxIndex),
		anomalyCell(p.TxHash),
		anomalyCell(p.Detail),
	)
	if _, writeErr := file.WriteString(row); writeErr != nil {
		return fmt.Errorf("write anomaly row: %w", writeErr)
	}
	return nil
}

// Build anomaly records for an empty attestor field in a successful eth_getBlockByNumber response.
func emptyAttestorRecords(blockNumber *big.Int, blockRaw json.RawMessage) ([]*RecordAnomalyParams, error) {
	if len(blockRaw) == 0 || string(blockRaw) == "null" {
		return nil, nil
	}
	found, value, err := rpc.DetectEmptyAttestor(blockRaw)
	if err != nil || !found {
		return nil, err
	}
	return []*RecordAnomalyParams{{
		Type:        AnomalyTypeEmptyAttestor,
		BlockNumber: blockNumber.Text(10),
		TxIndex:     -1,
		Detail:      fmt.Sprintf("attestor field is %s", value),
	}}, nil
}

// Build anomaly records for transactions reported as errors inside a successful debug_traceBlockByNumber response.
func traceTxErrorRecords(blockNumber *big.Int, traceRaw json.RawMessage) ([]*RecordAnomalyParams, error) {
	if len(traceRaw) == 0 || string(traceRaw) == "null" {
		return nil, nil
	}
	txErrors, err := rpc.DetectTraceTxErrors(traceRaw)
	if err != nil {
		return nil, err
	}
	records := make([]*RecordAnomalyParams, 0, len(txErrors))
	for _, txError := range txErrors {
		records = append(records, &RecordAnomalyParams{
			Type:        AnomalyTypeTraceTxError,
			BlockNumber: blockNumber.Text(10),
			TxIndex:     txError.Index,
			TxHash:      txError.TxHash,
			Detail:      txError.Detail,
		})
	}
	return records, nil
}

// Render a transaction index cell; negative index renders empty.
func txIndexCell(txIndex int) string {
	if txIndex < 0 {
		return ""
	}
	return strconv.Itoa(txIndex)
}
