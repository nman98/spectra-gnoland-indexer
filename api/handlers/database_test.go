package handlers_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

type MockDatabase struct {
	blocks       map[uint64]*database.BlockData
	transactions map[string]*database.Transaction
	addressTxs   map[string]*[]database.AddressTx
	blockSigners map[uint64]*database.BlockSigners
	latestBlock  *database.BlockData

	bankSend      map[string]*database.BankSend
	msgCall       map[string]*database.MsgCall
	msgAddPackage map[string]*database.MsgAddPackage
	msgRun        map[string]*database.MsgRun
	msgTypes      map[string][]string

	shouldError bool
	errorMsg    string
	// notFoundError toggles the simulated failure mode: when true, the mock
	// wraps errorMsg with database.ErrNotFound so handlers should translate it
	// into a 404. When false, the mock returns a plain error so handlers
	// should mask it behind a generic 500.
	notFoundError bool
}

// simulatedError builds the error the mock should return when shouldError is
// set. It honours notFoundError so tests can exercise both the 404 and 500
// paths in the handlers.
func (m *MockDatabase) simulatedError() error {
	if m.notFoundError {
		return fmt.Errorf("%s: %w", m.errorMsg, database.ErrNotFound)
	}
	return fmt.Errorf("%s", m.errorMsg)
}

// notFoundErr wraps a descriptive message with database.ErrNotFound so the
// mock mimics the sentinel the real queries return when a lookup legitimately
// misses.
func notFoundErr(msg string) error {
	return fmt.Errorf("%s: %w", msg, database.ErrNotFound)
}

func (m *MockDatabase) GetBlock(ctx context.Context, height uint64, chainName string) (*database.BlockData, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	block, ok := m.blocks[height]
	if !ok {
		return nil, notFoundErr("block not found")
	}
	return block, nil
}

func (m *MockDatabase) GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	var result []*database.BlockData
	for i := fromHeight; i <= toHeight; i++ {
		block, ok := m.blocks[i]
		if !ok {
			return nil, notFoundErr("block not found")
		}
		result = append(result, block)
	}
	return result, nil
}

func (m *MockDatabase) GetAllBlockSigners(ctx context.Context, chainName string, blockHeight uint64) (*database.BlockSigners, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	blockSigners, ok := m.blockSigners[blockHeight]
	if !ok {
		return nil, notFoundErr("block signers not found")
	}
	return blockSigners, nil
}

func (m *MockDatabase) GetTransaction(ctx context.Context, txHash string, chainName string) (*database.Transaction, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	transaction, ok := m.transactions[txHash]
	if !ok {
		return nil, notFoundErr("transaction not found")
	}
	return transaction, nil
}

func (m *MockDatabase) GetAddressTxs(
	ctx context.Context,
	address string,
	chainName string,
	fromTimestamp *time.Time,
	toTimestamp *time.Time,
	limit *uint64,
	cursor *string,
	direction database.Direction,
	sortOrder database.SortOrder,
) (*[]database.AddressTx, bool, error) {
	if m.shouldError {
		return nil, false, m.simulatedError()
	}
	addressTxs, ok := m.addressTxs[address]
	if !ok {
		return nil, false, notFoundErr("address transactions not found")
	}
	return addressTxs, false, nil
}

func (m *MockDatabase) GetLatestBlock(ctx context.Context, chainName string) (*database.BlockData, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	if m.latestBlock == nil {
		return nil, notFoundErr("latest block not found")
	}
	return m.latestBlock, nil
}

func (m *MockDatabase) GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*database.BlockData, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	blocks := make([]*database.BlockData, 0, len(m.blocks))
	for _, block := range m.blocks {
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (m *MockDatabase) GetLastXTransactions(
	ctx context.Context, chainName string, x uint64, sortOrder *database.SortOrder,
) ([]*database.Transaction, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	transactions := make([]*database.Transaction, 0, len(m.transactions))
	for _, transaction := range m.transactions {
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (m *MockDatabase) GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	msgTypes, ok := m.msgTypes[txHash]
	if !ok {
		return nil, notFoundErr("message type not found")
	}
	return msgTypes, nil
}

func (m *MockDatabase) GetBankSend(ctx context.Context, txHash string, chainName string) ([]*database.BankSend, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	bankSend, ok := m.bankSend[txHash]
	if !ok {
		return nil, notFoundErr("bank send not found")
	}
	return []*database.BankSend{bankSend}, nil
}

func (m *MockDatabase) GetMsgCall(ctx context.Context, txHash string, chainName string) ([]*database.MsgCall, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	msgCall, ok := m.msgCall[txHash]
	if !ok {
		return nil, notFoundErr("message call not found")
	}
	return []*database.MsgCall{msgCall}, nil
}

func (m *MockDatabase) GetMsgAddPackage(ctx context.Context, txHash string, chainName string) ([]*database.MsgAddPackage, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	msgAddPackage, ok := m.msgAddPackage[txHash]
	if !ok {
		return nil, notFoundErr("message add package not found")
	}
	return []*database.MsgAddPackage{msgAddPackage}, nil
}

func (m *MockDatabase) GetMsgRun(ctx context.Context, txHash string, chainName string) ([]*database.MsgRun, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	msgRun, ok := m.msgRun[txHash]
	if !ok {
		return nil, notFoundErr("message run not found")
	}
	return []*database.MsgRun{msgRun}, nil
}

func (m *MockDatabase) GetTransactionsByRange(
	ctx context.Context, chainName string, cursor string, limit uint64, direction database.Direction,
) ([]*database.Transaction, bool, error) {
	if m.shouldError {
		return nil, false, m.simulatedError()
	}
	transactions := make([]*database.Transaction, 0, len(m.transactions))
	for _, transaction := range m.transactions {
		transactions = append(transactions, transaction)
	}
	return transactions, false, nil
}

func (m *MockDatabase) GetTotalTxCount24h(ctx context.Context, chainName string) (int64, error) {
	if m.shouldError {
		return 0, m.simulatedError()
	}
	return int64(len(m.transactions)), nil
}

func (m *MockDatabase) GetTotalTxCountByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) ([]*database.TxCountDateRange, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	return []*database.TxCountDateRange{}, nil
}

func (m *MockDatabase) GetTotalTxCountByHour(
	ctx context.Context,
	chainName string,
	fromTimestamp time.Time,
	toTimestamp time.Time,
	sortOrder database.SortOrder,
) ([]*database.TxCountTimeRange, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	return []*database.TxCountTimeRange{}, nil
}

func (m *MockDatabase) GetVolumeByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) (database.VolumeByDenomDaily, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	return database.VolumeByDenomDaily{}, nil
}

func (m *MockDatabase) GetVolumeByHour(
	ctx context.Context,
	chainName string,
	fromTimestamp time.Time,
	toTimestamp time.Time,
	sortOrder database.SortOrder,
) (database.VolumeByDenomHourly, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	return database.VolumeByDenomHourly{}, nil
}

func (m *MockDatabase) GetBlockCount24h(ctx context.Context, chainName string) (int64, error) {
	if m.shouldError {
		return 0, m.simulatedError()
	}
	return int64(len(m.blocks)), nil
}

func (m *MockDatabase) GetBlockCountByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) ([]*database.BlockCountByDate, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	return []*database.BlockCountByDate{}, nil
}

func (m *MockDatabase) GetDailyActiveAccount(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) ([]*database.DailyActiveAccount, error) {
	if m.shouldError {
		return nil, m.simulatedError()
	}
	return []*database.DailyActiveAccount{}, nil
}
