package timescaledb

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/jackc/pgx/v5"
)

// GetBankSend returns the bank send messages for a given transaction hash.
func (t *TimescaleDb) GetBankSend(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.BankSend, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
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
	JOIN tx_hash_id id ON bms.tx_id = id.tx_id AND bms.chain_name = id.chain_name
	LEFT JOIN gno_addresses gn_from ON bms.from_address = gn_from.id
	LEFT JOIN gno_addresses gn_to ON bms.to_address = gn_to.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND bms.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bankSends := make([]*database.BankSend, 0)
	for rows.Next() {
		bankSend := &database.BankSend{}
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

// GetMsgCall returns the vm_msg_call messages for a given transaction hash.
func (t *TimescaleDb) GetMsgCall(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.MsgCall, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
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
	JOIN tx_hash_id id ON vmc.tx_id = id.tx_id AND vmc.chain_name = id.chain_name
	LEFT JOIN gno_addresses gn ON vmc.caller = gn.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND vmc.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgCalls := make([]*database.MsgCall, 0)
	for rows.Next() {
		msgCall := &database.MsgCall{}
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

// GetMsgAddPackage returns the vm_msg_add_package messages for a given transaction hash.
func (t *TimescaleDb) GetMsgAddPackage(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.MsgAddPackage, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
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
	JOIN tx_hash_id id ON vmap.tx_id = id.tx_id AND vmap.chain_name = id.chain_name
	LEFT JOIN gno_addresses gn ON vmap.creator = gn.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND vmap.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgAddPackages := make([]*database.MsgAddPackage, 0)
	for rows.Next() {
		msgAddPackage := &database.MsgAddPackage{}
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

// GetMsgRun returns the vm_msg_run messages for a given transaction hash.
func (t *TimescaleDb) GetMsgRun(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.MsgRun, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
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
	JOIN tx_hash_id id ON vmr.tx_id = id.tx_id AND vmr.chain_name = id.chain_name
	LEFT JOIN gno_addresses gn ON vmr.caller = gn.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND vmr.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgRuns := make([]*database.MsgRun, 0)
	for rows.Next() {
		msgRun := &database.MsgRun{}
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

// GetBankMultiSend returns the bank multi-send message rows for a given transaction hash.
// Each row represents one input (direction=false) or output (direction=true) entry.
func (t *TimescaleDb) GetBankMultiSend(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.BankMultiSendRow, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
	bms.message_counter,
	bms.timestamp,
	bms.direction,
	gna.address,
	bms.coins,
	array(
		SELECT gn.address
		FROM unnest(bms.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM bank_msg_multi_send bms
	JOIN tx_hash_id id ON bms.tx_id = id.tx_id AND bms.chain_name = id.chain_name
	LEFT JOIN gno_addresses gna ON bms.address_id = gna.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND bms.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*database.BankMultiSendRow, 0)
	for rows.Next() {
		row := &database.BankMultiSendRow{}
		err := rows.Scan(
			&row.TxHash,
			&row.MessageCounter,
			&row.Timestamp,
			&row.Direction,
			&row.Address,
			&row.Coins,
			&row.Signers,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMsgAuthCrSession returns the auth create-session messages for a given transaction hash.
func (t *TimescaleDb) GetMsgAuthCrSession(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.MsgAuthCrSession, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
	acs.message_counter,
	acs.timestamp,
	gn_creator.address AS creator,
	gn_session.address AS session_key,
	acs.expires_at,
	acs.spend_limit,
	acs.spend_period,
	acs.allow_paths,
	array(
		SELECT gn.address
		FROM unnest(acs.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM auth_msg_create_session acs
	JOIN tx_hash_id id ON acs.tx_id = id.tx_id AND acs.chain_name = id.chain_name
	LEFT JOIN gno_addresses gn_creator ON acs.creator = gn_creator.id
	LEFT JOIN gno_addresses gn_session ON acs.session_key = gn_session.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND acs.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*database.MsgAuthCrSession, 0)
	for rows.Next() {
		msg := &database.MsgAuthCrSession{}
		err := rows.Scan(
			&msg.TxHash,
			&msg.MessageCounter,
			&msg.Timestamp,
			&msg.Creator,
			&msg.SessionKey,
			&msg.ExpiresAt,
			&msg.SpendLimit,
			&msg.SpendPeriod,
			&msg.AllowPaths,
			&msg.Signers,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMsgAuthRvSession returns the auth revoke-session messages for a given transaction hash.
func (t *TimescaleDb) GetMsgAuthRvSession(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.MsgAuthRvSession, error) {
	query := `
	SELECT
    encode(id.tx_hash, 'base64') AS tx_hash,
    rvs.message_counter,
    rvs.timestamp,
    gn_creator.address AS creator,
    gn_session.address AS session_key,
    array(
        SELECT gn.address
        FROM unnest(rvs.signers) AS signer_id
        JOIN gno_addresses gn ON gn.id = signer_id
    ) AS signers
    FROM auth_msg_revoke_session rvs
    JOIN tx_hash_id id ON rvs.tx_id = id.tx_id AND rvs.chain_name = id.chain_name
    LEFT JOIN gno_addresses gn_creator ON rvs.creator = gn_creator.id
    LEFT JOIN gno_addresses gn_session ON rvs.session_key = gn_session.id
    WHERE id.tx_hash = decode($1, 'base64')
    AND rvs.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*database.MsgAuthRvSession, 0)
	for rows.Next() {
		msg := &database.MsgAuthRvSession{}
		err := rows.Scan(
			&msg.TxHash,
			&msg.MessageCounter,
			&msg.Timestamp,
			&msg.Creator,
			&msg.SessionKey,
			&msg.Signers,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMsgAuthRvAllSessions returns the auth revoke-all-sessions messages for a given transaction hash.
func (t *TimescaleDb) GetMsgAuthRvAllSessions(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*database.MsgAuthRvAllSessions, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
	rvas.message_counter,
	rvas.timestamp,
	gn.address AS creator,
	array(
		SELECT gn.address
		FROM unnest(rvas.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM auth_msg_revoke_all_sessions rvas
	JOIN tx_hash_id id ON rvas.tx_id = id.tx_id AND rvas.chain_name = id.chain_name
	LEFT JOIN gno_addresses gn ON rvas.creator = gn.id
	WHERE id.tx_hash = decode($1, 'base64')
	AND rvas.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*database.MsgAuthRvAllSessions, 0)
	for rows.Next() {
		msg := &database.MsgAuthRvAllSessions{}
		err := rows.Scan(
			&msg.TxHash,
			&msg.MessageCounter,
			&msg.Timestamp,
			&msg.Creator,
			&msg.Signers,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMsgTypes returns the message types for a given transaction hash.
func (t *TimescaleDb) GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error) {
	query := `
	SELECT tg.msg_types
	FROM transaction_general tg
	JOIN tx_hash_id id ON tg.tx_id = id.tx_id AND tg.chain_name = id.chain_name
	WHERE id.tx_hash = decode($1, 'base64')
	AND tg.chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var msgTypes []string
	err := row.Scan(&msgTypes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("msg types for tx %s: %w", txHash, database.ErrNotFound)
		}
		return nil, err
	}
	return msgTypes, nil
}
