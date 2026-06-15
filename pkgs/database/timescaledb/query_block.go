package timescaledb

import (
	"context"
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"golang.org/x/sync/errgroup"
)

// GetBlock gets a block from the database for a given height and chain name.
func (t *TimescaleDb) GetBlock(ctx context.Context, height uint64, chainName string) (*database.BlockData, error) {
	query1 := `
	SELECT encode(hash, 'base64'),
	b.height as height,
	b.timestamp as timestamp,
	b.chain_id as chain_id,
	gv.address as proposer
	FROM blocks b
	JOIN validator_block_signing vb ON b.height = vb.block_height AND b.chain_name = vb.chain_name
	JOIN gno_validators gv ON vb.proposer = gv.id
	WHERE b.height = $1
	AND b.chain_name = $2
	`
	query2 := `
	SELECT
	encode(id.tx_hash, 'base64'),
	tg.block_height
	FROM transaction_general tg
	JOIN tx_hash_id id ON tg.tx_id = id.tx_id AND tg.chain_name = id.chain_name
	WHERE tg.chain_name = $1
	AND tg.block_height = $2
	`
	var blocks []*database.BlockData
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
		return nil, fmt.Errorf("block at height %d: %w", height, database.ErrNotFound)
	}

	if txs[blocks[0].Height] != nil {
		blocks[0].Txs = append(blocks[0].Txs, txs[blocks[0].Height]...)
	}
	blocks[0].TxCounter = len(blocks[0].Txs)
	return blocks[0], nil
}

func (t *TimescaleDb) GetLatestBlock(ctx context.Context, chainName string) (*database.BlockData, error) {
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
	encode(id.tx_hash, 'base64'),
	tg.block_height
	FROM transaction_general tg
	JOIN tx_hash_id id ON tg.tx_id = id.tx_id AND tg.chain_name = id.chain_name
	WHERE tg.chain_name = $1
	AND tg.block_height = (SELECT MAX(height) FROM blocks WHERE chain_name = $1)
	`
	var blocks []*database.BlockData
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
		return nil, fmt.Errorf("no blocks for chain %q: %w", chainName, database.ErrNotFound)
	}

	if txs[blocks[0].Height] != nil {
		blocks[0].Txs = append(blocks[0].Txs, txs[blocks[0].Height]...)
	}
	blocks[0].TxCounter = len(blocks[0].Txs)
	return blocks[0], nil
}

// GetLastXBlocks gets the last x blocks for a given chain name.
func (t *TimescaleDb) GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*database.BlockData, error) {
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
	encode(id.tx_hash, 'base64'),
	tg.block_height
	FROM transaction_general tg
	JOIN tx_hash_id id ON tg.tx_id = id.tx_id AND tg.chain_name = id.chain_name
	WHERE tg.chain_name = $1
	AND tg.block_height <= (SELECT MAX(height) FROM blocks WHERE chain_name = $1)
	AND tg.block_height >= (SELECT MAX(height) FROM blocks WHERE chain_name = $1) - $2
	ORDER BY tg.block_height DESC
	`

	var blocks []*database.BlockData
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
		return nil, fmt.Errorf("no blocks for chain %q: %w", chainName, database.ErrNotFound)
	}
	for _, block := range blocks {
		if txs[block.Height] != nil {
			block.Txs = append(block.Txs, txs[block.Height]...)
		}
		block.TxCounter = len(block.Txs)
	}
	return blocks, nil
}

// GetFromToBlocks gets a range of blocks between fromHeight and toHeight (inclusive).
func (t *TimescaleDb) GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*database.BlockData, error) {
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
	encode(id.tx_hash, 'base64'),
	tg.block_height
	FROM transaction_general tg
	JOIN tx_hash_id id ON tg.tx_id = id.tx_id AND tg.chain_name = id.chain_name
	WHERE tg.chain_name = $1
	AND tg.block_height >= $2 AND tg.block_height <= $3
	`
	var blocks []*database.BlockData
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
			fromHeight, toHeight, chainName, database.ErrNotFound,
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
	query := `
	SELECT height, timestamp
	FROM blocks
	WHERE chain_name = $1
	AND height IN (
	    (SELECT MAX(height) FROM blocks WHERE chain_name = $1),
	    (SELECT MIN(height) FROM blocks WHERE chain_name = $1)
	)
	ORDER BY height DESC
	`

	rows, err := t.pool.Query(ctx, query, chainName)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type blockPoint struct {
		height uint64
		ts     time.Time
	}
	var points []blockPoint

	for rows.Next() {
		var p blockPoint
		if err := rows.Scan(&p.height, &p.ts); err != nil {
			return 0, err
		}
		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(points) < 2 {
		return 0, fmt.Errorf("not enough block data to compute average block time for chain %q", chainName)
	}

	diffHeight := points[0].height - points[1].height
	if diffHeight == 0 {
		return 0, fmt.Errorf("not enough block data to compute average block time for chain %q", chainName)
	}

	return points[0].ts.Sub(points[1].ts).Seconds() / float64(diffHeight), nil
}

func (t *TimescaleDb) fetchBlocksData(ctx context.Context, query string, args ...any) ([]*database.BlockData, error) {
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocks := make([]*database.BlockData, 0)
	for rows.Next() {
		block := &database.BlockData{}
		err := rows.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Proposer)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func (t *TimescaleDb) fetchTransactionData(ctx context.Context, query string, args ...any) (map[uint64][]string, error) {
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs := make(map[uint64][]string)
	for rows.Next() {
		tx := &database.Transaction{}
		err := rows.Scan(&tx.TxHash, &tx.BlockHeight)
		if err != nil {
			return nil, err
		}
		txs[tx.BlockHeight] = append(txs[tx.BlockHeight], tx.TxHash)
	}
	return txs, rows.Err()
}
