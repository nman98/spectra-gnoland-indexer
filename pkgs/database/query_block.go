package database

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

// GetBlock gets a block from the database for a given height and chain name
//
// Usage:
//
// # Used to get a block from the database for a given height and chain name
//
// Parameters:
//   - height: the height of the block
//   - chainName: the name of the chain
//
// Returns:
//   - *BlockData: the block data
//   - error: if the query fails
func (t *TimescaleDb) GetBlock(ctx context.Context, height uint64, chainName string) (*BlockData, error) {
	query1 := `
	SELECT encode(hash, 'base64'), 
	height, 
	timestamp, 
	chain_id
	FROM blocks
	WHERE height = $1
	AND chain_name = $2
	`
	query2 := `
	SELECT
	encode(tx_hash, 'base64'),
	block_height
	FROM transaction_general
	WHERE chain_name = $1
	AND block_height = $2
	`
	var blocks []*BlockData
	var txs map[uint64][]string

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		blocks, err = t.fetchBlocksData(ctx, query1, height, chainName)
		return err
	})

	eg.Go(func() error {
		var err error
		txs, err = t.fetchTransactionData(ctx, query2, chainName, height)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("block at height %d: %w", height, ErrNotFound)
	}

	if txs[blocks[0].Height] != nil {
		blocks[0].Txs = append(blocks[0].Txs, txs[blocks[0].Height]...)
	}
	blocks[0].TxCounter = len(blocks[0].Txs)
	return blocks[0], nil
}

func (t *TimescaleDb) GetLatestBlock(ctx context.Context, chainName string) (*BlockData, error) {
	query1 := `
	SELECT encode(hash, 'base64'), 
	b.height as height, 
	b.timestamp as timestamp, 
	b.chain_id as chain_id,
	gv.address as proposer
	FROM blocks b
	JOIN validator_block_signing vb ON b.height = vb.block_height AND b.chain_name = vb.chain_name
	JOIN gno_validators gv ON vb.proposer = gv.id
	WHERE b.chain_name = $1
	ORDER BY b.height DESC
	LIMIT 1
	`
	query2 := `
	SELECT
	encode(tx_hash, 'base64'),
	block_height
	FROM transaction_general
	WHERE chain_name = $1
	AND block_height = (SELECT MAX(height) FROM blocks WHERE chain_name = $1)  
	`
	var blocks []*BlockData
	var txs map[uint64][]string

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		blocks, err = t.fetchBlocksData(ctx, query1, chainName)
		return err
	})

	eg.Go(func() error {
		var err error
		txs, err = t.fetchTransactionData(ctx, query2, chainName)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks for chain %q: %w", chainName, ErrNotFound)
	}

	if txs[blocks[0].Height] != nil {
		blocks[0].Txs = append(blocks[0].Txs, txs[blocks[0].Height]...)
	}
	blocks[0].TxCounter = len(blocks[0].Txs)
	return blocks[0], nil
}

// GetLastXBlocks gets the last x blocks from the database for a given chain name
//
// Usage:
//
// # Used to get the last x blocks from the database for a given chain name
//
// Parameters:
//   - chainName: the name of the chain
//   - x: the number of blocks to get
//
// Returns:
//   - []*BlockData: the last x blocks
//   - error: if the query fails
func (t *TimescaleDb) GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*BlockData, error) {
	query1 := `
	SELECT encode(hash, 'base64'), 
	b.height as height, 
	b.timestamp as timestamp, 
	b.chain_id as chain_id,
	gv.address as proposer
	FROM blocks b
	JOIN validator_block_signing vb ON b.height = vb.block_height AND b.chain_name = vb.chain_name
	JOIN gno_validators gv ON vb.proposer = gv.id
	WHERE b.chain_name = $1
	ORDER BY b.height DESC
	LIMIT $2
	`
	query2 := `
	SELECT 
	encode(tx_hash, 'base64'),
	block_height
	FROM transaction_general
	WHERE chain_name = $1
	AND block_height <= (SELECT MAX(height) FROM blocks WHERE chain_name = $1) 
	AND block_height >= (SELECT MAX(height) FROM blocks WHERE chain_name = $1) - $2
	ORDER BY block_height DESC
	`

	var blocks []*BlockData
	var txs map[uint64][]string

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		blocks, err = t.fetchBlocksData(ctx, query1, chainName, x)
		return err
	})

	eg.Go(func() error {
		var err error
		txs, err = t.fetchTransactionData(ctx, query2, chainName, x)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks for chain %q: %w", chainName, ErrNotFound)
	}
	for _, block := range blocks {
		if txs[block.Height] != nil {
			block.Txs = append(block.Txs, txs[block.Height]...)
		}
		block.TxCounter = len(block.Txs)
	}
	return blocks, nil
}

// GetFromToBlocks gets a range of blocks from the database for a given height range and chain name
//
// Usage:
//
// # Used to get a range of blocks from the database for a given height range and chain name
//
// Parameters:
//   - fromHeight: the starting height of the block
//   - toHeight: the ending height of the block (inclusive)
//   - chainName: the name of the chain
//
// Returns:
//   - []*BlockData: the range of block data
//   - error: if the query fails
func (t *TimescaleDb) GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*BlockData, error) {
	query1 := `
	SELECT encode(hash, 'base64'), 
	b.height as height, 
	b.timestamp as timestamp, 
	b.chain_id as chain_id,
	gv.address as proposer
	FROM blocks b
	JOIN validator_block_signing vb ON b.height = vb.block_height AND b.chain_name = vb.chain_name
	JOIN gno_validators gv ON vb.proposer = gv.id
	WHERE b.height >= $1 AND b.height <= $2
	AND b.chain_name = $3
	`

	query2 := `
	SELECT
	encode(tx_hash, 'base64'),
	block_height
	FROM transaction_general
	WHERE chain_name = $1
	AND block_height >= $2 AND block_height <= $3
	`
	var blocks []*BlockData
	var txs map[uint64][]string

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		blocks, err = t.fetchBlocksData(ctx, query1, fromHeight, toHeight, chainName)
		return err
	})

	eg.Go(func() error {
		var err error
		txs, err = t.fetchTransactionData(ctx, query2, chainName, fromHeight, toHeight)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf(
			"no blocks between heights %d and %d for chain %q: %w",
			fromHeight, toHeight, chainName, ErrNotFound,
		)
	}

	for _, block := range blocks {
		if txs[block.Height] != nil {
			block.Txs = append(block.Txs, txs[block.Height]...)
		}
		block.TxCounter = len(block.Txs)
	}

	return blocks, nil
}

func (t *TimescaleDb) GetAvgBlockProdTime(ctx context.Context, chainName string) (float64, error) {
	// get latest block height
	blockData, err := t.GetLatestBlock(ctx, chainName)
	if err != nil {
		return 0, err
	}

	// compare latest block height with the height - 10K from the latest
	latestHeight := blockData.Height
	var compHeight uint64
	var diffHeight uint64 = 10000
	if latestHeight <= 10000 {
		compHeight = 1
		diffHeight = latestHeight - compHeight
	} else {
		compHeight = latestHeight - 10000
	}

	query := `
	SELECT
	timestamp
	FROM blocks
	WHERE height IN ($1, $2) AND chain_name = $3
	ORDER BY height DESC
	`

	rows, err := t.pool.Query(ctx, query, latestHeight, compHeight, chainName)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var timestamps []time.Time
	for rows.Next() {
		var timestamp time.Time
		err := rows.Scan(&timestamp)
		if err != nil {
			return 0, err
		}
		timestamps = append(timestamps, timestamp)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(timestamps) != 2 {
		return 0, fmt.Errorf("expected 2 timestamps, got %d", len(timestamps))
	}
	latestTimestamp := timestamps[0]
	compTimestamp := timestamps[1]

	calc := (latestTimestamp.Sub(compTimestamp) / time.Duration(diffHeight)).Seconds()

	return calc, nil
}

// fetchBlocksData fetches block data from the database and appends to the provided slice
func (t *TimescaleDb) fetchBlocksData(ctx context.Context, query string, args ...interface{}) ([]*BlockData, error) {
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocks := make([]*BlockData, 0)
	for rows.Next() {
		block := &BlockData{}
		err := rows.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Proposer)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

// fetchTransactionData fetches transaction data and maps it by block height
func (t *TimescaleDb) fetchTransactionData(ctx context.Context, query string, args ...interface{}) (map[uint64][]string, error) {
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs := make(map[uint64][]string)
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(&tx.TxHash, &tx.BlockHeight)
		if err != nil {
			return nil, err
		}
		txs[tx.BlockHeight] = append(txs[tx.BlockHeight], tx.TxHash)
	}
	return txs, rows.Err()
}
