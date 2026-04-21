package handlers

import (
	"context"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

type AddressDbHandler interface {
	GetAddressTxs(
		ctx context.Context,
		address string,
		chainName string,
		fromTimestamp *time.Time,
		toTimestamp *time.Time,
		limit *uint64,
		cursor *string,
		direction database.Direction,
		sortOrder database.SortOrder,
	) (*[]database.AddressTx, bool, error)
	GetDailyActiveAccount(
		ctx context.Context,
		chainName string,
		dateFrom date.Date,
		dateTo date.Date,
		sortOrder database.SortOrder,
	) ([]*database.DailyActiveAccount, error)
}

type ValidatorDbHandler interface {
	GetValidatorSigning24h(
		ctx context.Context,
		validatorAddress string,
		chainName string,
	) (*database.ValidatorSigning, error)
	GetValidatorSigningByHour(
		ctx context.Context,
		validatorAddress string,
		chainName string,
		fromTimestamp time.Time,
		toTimestamp time.Time,
		sortOrder database.SortOrder,
	) ([]*database.ValidatorSigning, error)
	GetAllValidators(ctx context.Context, chainName string) (*database.ValidatorList, error)
}

type BlockDbHandler interface {
	GetBlock(ctx context.Context, height uint64, chainName string) (*database.BlockData, error)
	GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error)
	GetAllBlockSigners(ctx context.Context, chainName string, blockHeight uint64) (*database.BlockSigners, error)
	GetLatestBlock(ctx context.Context, chainName string) (*database.BlockData, error)
	GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*database.BlockData, error)
	GetBlockCount24h(ctx context.Context, chainName string) (int64, error)
	GetBlockCountByDate(
		ctx context.Context,
		chainName string,
		dateFrom date.Date,
		dateTo date.Date,
		sortOrder database.SortOrder,
	) ([]*database.BlockCountByDate, error)
}

type TransactionDbHandler interface {
	GetTransaction(ctx context.Context, txHash string, chainName string) (*database.Transaction, error)
	GetLastXTransactions(ctx context.Context, chainName string, x uint64, sortOrder *database.SortOrder,
	) ([]*database.Transaction, error)
	GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error)
	GetBankSend(ctx context.Context, txHash string, chainName string) ([]*database.BankSend, error)
	GetMsgCall(ctx context.Context, txHash string, chainName string) ([]*database.MsgCall, error)
	GetMsgAddPackage(ctx context.Context, txHash string, chainName string) ([]*database.MsgAddPackage, error)
	GetMsgRun(ctx context.Context, txHash string, chainName string) ([]*database.MsgRun, error)
	GetTransactionsByRange(
		ctx context.Context, chainName string, cursor string, limit uint64, direction database.Direction,
	) ([]*database.Transaction, bool, error)
	GetTotalTxCount24h(ctx context.Context, chainName string) (int64, error)
	GetTotalTxCountByDate(ctx context.Context, chainName string, dateFrom date.Date, dateTo date.Date, sortOrder database.SortOrder) ([]*database.TxCountDateRange, error)
	GetTotalTxCountByHour(ctx context.Context, chainName string, fromTimestamp time.Time, toTimestamp time.Time, sortOrder database.SortOrder) ([]*database.TxCountTimeRange, error)
	GetVolumeByDate(ctx context.Context, chainName string, dateFrom date.Date, dateTo date.Date, sortOrder database.SortOrder) (database.VolumeByDenomDaily, error)
	GetVolumeByHour(ctx context.Context, chainName string, fromTimestamp time.Time, toTimestamp time.Time, sortOrder database.SortOrder) (database.VolumeByDenomHourly, error)
}

type InMemoryDbHandler interface {
	GetTotalAddressesCount(ctx context.Context, chainName string) (int32, error)
	GetAvgBlockProdTime(ctx context.Context, chainName string) (float64, error)
}
