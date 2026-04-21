package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// GetAllBlockSigners gets all of the validators that signed that block + the proposer
//
// Usage:
//
// # Used to get all of the validators that signed that block + the proposer
//
// Parameters:
//   - chainName: the name of the chain
//   - blockHeight: the height of the block
//
// Returns:
//   - *BlockSigners: the block signers
//   - error: if the query fails
func (t *TimescaleDb) GetAllBlockSigners(
	ctx context.Context,
	chainName string,
	blockHeight uint64,
) (*BlockSigners, error) {
	query := `
	SELECT
	vb.block_height,
	gv.address AS proposer,
	array(
		SELECT gv.address 
		FROM unnest(vb.signed_vals) AS signed_val_id
		JOIN gno_validators gv ON gv.id = signed_val_id
	) AS signed_vals
	FROM validator_block_signing vb
	LEFT JOIN gno_validators gv ON vb.proposer = gv.id
	WHERE vb.chain_name = $1
	AND vb.block_height = $2
	`
	row := t.pool.QueryRow(ctx, query, chainName, blockHeight)
	var blockSigners BlockSigners
	err := row.Scan(&blockSigners.BlockHeight, &blockSigners.Proposer, &blockSigners.SignedVals)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("block signers at height %d: %w", blockHeight, ErrNotFound)
		}
		return nil, err
	}
	return &blockSigners, nil
}

// GetBankSend gets the bank send message for a given transaction hash
//
// Usage:
//
// # Used to get the bank send message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*BankSend: the bank send messages
//   - error: if the query fails
func (t *TimescaleDb) GetBankSend(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*BankSend, error) {
	query := `
	SELECT 
    encode(bms.tx_hash, 'base64') AS tx_hash,
    bms.timestamp,
    gn_from.address AS from_address,
    gn_to.address AS to_address,
    bms.amount,
    array(
        SELECT gn.address 
        FROM unnest(bms.signers) AS signer_id
        JOIN gno_addresses gn ON gn.id = signer_id
    ) AS signers
	FROM bank_msg_send bms
	LEFT JOIN gno_addresses gn_from ON bms.from_address = gn_from.id
	LEFT JOIN gno_addresses gn_to ON bms.to_address = gn_to.id
	WHERE bms.tx_hash = decode($1, 'base64')
	AND bms.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bankSends := make([]*BankSend, 0)
	for rows.Next() {
		bankSend := &BankSend{}
		err := rows.Scan(
			&bankSend.TxHash,
			&bankSend.Timestamp,
			&bankSend.FromAddress,
			&bankSend.ToAddress,
			&bankSend.Amount,
			&bankSend.Signers,
		)
		if err != nil {
			return nil, err
		}
		bankSends = append(bankSends, bankSend)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bankSends, nil
}

// GetMsgCall gets the msg call message for a given transaction hash
//
// Usage:
//
// # Used to get the msg call message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*MsgCall: the msg call messages
//   - error: if the query fails
func (t *TimescaleDb) GetMsgCall(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*MsgCall, error) {
	query := `
	SELECT 
	encode(vmc.tx_hash, 'base64') AS tx_hash,
	vmc.message_counter,
	vmc.timestamp,
	gn.address AS caller,
	vmc.pkg_path,
	vmc.func_name,
	vmc.args,
	vmc.send,
	vmc.max_deposit,
	array(
		SELECT gn.address 
		FROM unnest(vmc.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM vm_msg_call vmc
	LEFT JOIN gno_addresses gn ON vmc.caller = gn.id
	WHERE vmc.tx_hash = decode($1, 'base64')
	AND vmc.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgCalls := make([]*MsgCall, 0)
	for rows.Next() {
		msgCall := &MsgCall{}
		err := rows.Scan(
			&msgCall.TxHash,
			&msgCall.MessageCounter,
			&msgCall.Timestamp,
			&msgCall.Caller,
			&msgCall.PkgPath,
			&msgCall.FuncName,
			&msgCall.Args,
			&msgCall.Send,
			&msgCall.MaxDeposit,
			&msgCall.Signers,
		)
		if err != nil {
			return nil, err
		}
		msgCalls = append(msgCalls, msgCall)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgCalls, nil
}

// GetMsgAddPackage gets the msg add package message for a given transaction hash
//
// Usage:
//
// # Used to get the msg add package message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*MsgAddPackage: the msg add package messages
//   - error: if the query fails
func (t *TimescaleDb) GetMsgAddPackage(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*MsgAddPackage, error) {
	query := `
	SELECT 
	encode(vmap.tx_hash, 'base64') AS tx_hash,
	vmap.message_counter,
	vmap.timestamp,
	gn.address AS creator,
	vmap.pkg_path,
	vmap.pkg_name,
	vmap.pkg_file_names,
	vmap.send,
	vmap.max_deposit,
	array(
		SELECT gn.address 
		FROM unnest(vmap.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM vm_msg_add_package vmap
	LEFT JOIN gno_addresses gn ON vmap.creator = gn.id
	WHERE vmap.tx_hash = decode($1, 'base64')
	AND vmap.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgAddPackages := make([]*MsgAddPackage, 0)
	for rows.Next() {
		msgAddPackage := &MsgAddPackage{}
		err := rows.Scan(
			&msgAddPackage.TxHash,
			&msgAddPackage.MessageCounter,
			&msgAddPackage.Timestamp,
			&msgAddPackage.Creator,
			&msgAddPackage.PkgPath,
			&msgAddPackage.PkgName,
			&msgAddPackage.PkgFileNames,
			&msgAddPackage.Send,
			&msgAddPackage.MaxDeposit,
			&msgAddPackage.Signers,
		)
		if err != nil {
			return nil, err
		}
		msgAddPackages = append(msgAddPackages, msgAddPackage)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgAddPackages, nil
}

// GetMsgRun gets the msg run message for a given transaction hash
//
// Usage:
//
// # Used to get the msg run message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*MsgRun: the msg run messages
//   - error: if the query fails
func (t *TimescaleDb) GetMsgRun(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*MsgRun, error) {
	query := `
	SELECT 
	encode(vmr.tx_hash, 'base64') AS tx_hash,
	vmr.message_counter,
	vmr.timestamp,
	gn.address AS caller,
	vmr.pkg_path,
	vmr.pkg_name,
	vmr.pkg_file_names,
	vmr.send,
	vmr.max_deposit,
	array(
		SELECT gn.address 
		FROM unnest(vmr.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM vm_msg_run vmr
	LEFT JOIN gno_addresses gn ON vmr.caller = gn.id
	WHERE vmr.tx_hash = decode($1, 'base64')
	AND vmr.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgRuns := make([]*MsgRun, 0)
	for rows.Next() {
		msgRun := &MsgRun{}
		err := rows.Scan(
			&msgRun.TxHash,
			&msgRun.MessageCounter,
			&msgRun.Timestamp,
			&msgRun.Caller,
			&msgRun.PkgPath,
			&msgRun.PkgName,
			&msgRun.PkgFileNames,
			&msgRun.Send,
			&msgRun.MaxDeposit,
			&msgRun.Signers,
		)
		if err != nil {
			return nil, err
		}
		msgRuns = append(msgRuns, msgRun)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgRuns, nil
}

// GetMsgTypes gets the message type for a given transaction hash
//
// Usage:
//
// # Used to get the message type for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []string: the message types
//   - error: if the query fails
func (t *TimescaleDb) GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error) {
	query := `
	SELECT msg_types
	FROM transaction_general
	WHERE tx_hash = decode($1, 'base64')
	AND chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var msgTypes []string
	err := row.Scan(&msgTypes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("msg types for tx %s: %w", txHash, ErrNotFound)
		}
		return nil, err
	}
	return msgTypes, nil
}
