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
}

func (m *MockDatabase) GetBlock(ctx context.Context, height uint64, chainName string) (*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	block, ok := m.blocks[height]
	if !ok {
		return nil, fmt.Errorf("block not found")
	}
	return block, nil
}

func (m *MockDatabase) GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	var result []*database.BlockData
	for i := fromHeight; i <= toHeight; i++ {
		block, ok := m.blocks[i]
		if !ok {
			return nil, fmt.Errorf("block not found")
		}
		result = append(result, block)
	}
	return result, nil
}

func (m *MockDatabase) GetAllBlockSigners(ctx context.Context, chainName string, blockHeight uint64) (*database.BlockSigners, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	blockSigners, ok := m.blockSigners[blockHeight]
	if !ok {
		return nil, fmt.Errorf("block signers not found")
	}
	return blockSigners, nil
}

func (m *MockDatabase) GetTransaction(ctx context.Context, txHash string, chainName string) (*database.Transaction, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	transaction, ok := m.transactions[txHash]
	if !ok {
		return nil, fmt.Errorf("transaction not found")
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
	page *uint64,
	cursor *string,
	sortOrder database.SortOrder,
) (*[]database.AddressTx, string, uint64, error) {
	if m.shouldError {
		return nil, "", 0, fmt.Errorf("%s", m.errorMsg)
	}
	addressTxs, ok := m.addressTxs[address]
	if !ok {
		return nil, "", 0, fmt.Errorf("address transactions not found")
	}
	txCount := uint64(len(*addressTxs))
	return addressTxs, "", txCount, nil
}

func (m *MockDatabase) GetLatestBlock(ctx context.Context, chainName string) (*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	if m.latestBlock == nil {
		return nil, fmt.Errorf("latest block not found")
	}
	return m.latestBlock, nil
}

func (m *MockDatabase) GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*database.BlockData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	transactions := make([]*database.Transaction, 0, len(m.transactions))
	for _, transaction := range m.transactions {
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (m *MockDatabase) GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgTypes, ok := m.msgTypes[txHash]
	if !ok {
		return nil, fmt.Errorf("message type not found")
	}
	return msgTypes, nil
}

func (m *MockDatabase) GetBankSend(ctx context.Context, txHash string, chainName string) ([]*database.BankSend, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	bankSend, ok := m.bankSend[txHash]
	if !ok {
		return nil, fmt.Errorf("bank send not found")
	}
	return []*database.BankSend{bankSend}, nil
}

func (m *MockDatabase) GetMsgCall(ctx context.Context, txHash string, chainName string) ([]*database.MsgCall, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgCall, ok := m.msgCall[txHash]
	if !ok {
		return nil, fmt.Errorf("message call not found")
	}
	return []*database.MsgCall{msgCall}, nil
}

func (m *MockDatabase) GetMsgAddPackage(ctx context.Context, txHash string, chainName string) ([]*database.MsgAddPackage, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgAddPackage, ok := m.msgAddPackage[txHash]
	if !ok {
		return nil, fmt.Errorf("message add package not found")
	}
	return []*database.MsgAddPackage{msgAddPackage}, nil
}

func (m *MockDatabase) GetMsgRun(ctx context.Context, txHash string, chainName string) ([]*database.MsgRun, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	msgRun, ok := m.msgRun[txHash]
	if !ok {
		return nil, fmt.Errorf("message run not found")
	}
	return []*database.MsgRun{msgRun}, nil
}

func (m *MockDatabase) GetTransactionsByCursor(
	ctx context.Context, chainName string, cursor string, limit uint64, sortOrder database.SortOrder,
) ([]*database.Transaction, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	transactions := make([]*database.Transaction, 0, len(m.transactions))
	for _, transaction := range m.transactions {
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (m *MockDatabase) GetTotalTxCount24h(ctx context.Context, chainName string) (int64, error) {
	if m.shouldError {
		return 0, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return database.VolumeByDenomHourly{}, nil
}

func (m *MockDatabase) GetBlockCount24h(ctx context.Context, chainName string) (int64, error) {
	if m.shouldError {
		return 0, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
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
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return []*database.DailyActiveAccount{}, nil
}
