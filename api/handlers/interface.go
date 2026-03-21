package handlers

import (
	"context"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

type AddressDbHandler interface {
	GetAddressTxs(
		ctx context.Context,
		address string,
		chainName string,
		fromTimestamp *time.Time,
		toTimestamp *time.Time,
		limit *uint64,
		page *uint64,
		cursor *string,
	) (*[]database.AddressTx, string, uint64, error)
	GetDailyActiveAccount(
		ctx context.Context,
		chainName string,
		date1 time.Time,
		date2 time.Time,
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
		date1 time.Time,
		date2 time.Time,
	) ([]*database.ValidatorSigning, error)
}

type BlockDbHandler interface {
	GetBlock(ctx context.Context, height uint64, chainName string) (*database.BlockData, error)
	GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error)
	GetAllBlockSigners(ctx context.Context, chainName string, blockHeight uint64) (*database.BlockSigners, error)
	GetLatestBlock(ctx context.Context, chainName string) (*database.BlockData, error)
	GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*database.BlockData, error)
	GetBlockCount24h(ctx context.Context, chainName string) (int64, error)
	GetBlockCountByDate(ctx context.Context, chainName string, date1 time.Time, date2 time.Time) ([]*database.BlockCountByDate, error)
}

type TransactionDbHandler interface {
	GetTransaction(ctx context.Context, txHash string, chainName string) (*database.Transaction, error)
	GetLastXTransactions(ctx context.Context, chainName string, x uint64) ([]*database.Transaction, error)
	GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error)
	GetBankSend(ctx context.Context, txHash string, chainName string) ([]*database.BankSend, error)
	GetMsgCall(ctx context.Context, txHash string, chainName string) ([]*database.MsgCall, error)
	GetMsgAddPackage(ctx context.Context, txHash string, chainName string) ([]*database.MsgAddPackage, error)
	GetMsgRun(ctx context.Context, txHash string, chainName string) ([]*database.MsgRun, error)
	GetTransactionsByCursor(ctx context.Context, chainName string, cursor string, limit uint64) ([]*database.Transaction, error)
	GetTotalTxCount24h(ctx context.Context, chainName string) (int64, error)
	GetTotalTxCountByDate(ctx context.Context, chainName string, date1 time.Time, date2 time.Time) ([]*database.TxCountTimeRange, error)
	GetTotalTxCountByHour(ctx context.Context, chainName string, date1 time.Time, date2 time.Time) ([]*database.TxCountTimeRange, error)
	GetVolumeByDate(ctx context.Context, chainName string, date1 time.Time, date2 time.Time) (database.VolumeByDenom, error)
	GetVolumeByHour(ctx context.Context, chainName string, date1 time.Time, date2 time.Time) (database.VolumeByDenom, error)
}

type InMemoryDbHandler interface {
	GetTotalAddressesCount(ctx context.Context, chainName string) (int32, error)
	GetAvgBlockProdTime(ctx context.Context, chainName string) (time.Duration, error)
}
