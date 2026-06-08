package rpcclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// NewRpcClient creates a new rpc client for the gnoland blockchain.
//
// This is the main entry point for the rpc client. It will create a new rpc client and return it.
// It has only methods that are used by the indexer.
// It is not recommended to use this client for other purposes unless the data is related to any kind of
// data analytics or storing data.
//
// Methods:
//
//   - Health: sync call to get the health of the rpc client
//   - GetValidators: call to get validators from the rpc client
//   - GetBlock: call to get a block from the rpc client
//   - GetTx: call to get a tx from the rpc client
//   - GetAbciQuery: call to get a abci query from the rpc client
//
// Parameters:
//
//   - rpcURL: the url of the rpc client
//   - timeout: the timeout for the rpc client(optional)
//
// Returns:
//
//   - *RpcGnoland: the rpc client
//   - error: if the rpc client fails to connect
func NewRpcClient(rpcURL string, timeout *time.Duration) (*RpcGnoland, error) {
	// validate the rpc url
	if rpcURL == "" {
		return nil, errors.New("rpcURL is required")
	} else if !strings.HasPrefix(rpcURL, "http://") && !strings.HasPrefix(rpcURL, "https://") {
		return nil, errors.New("rpcURL must start with http:// or https://")
	}
	// sanitize the rpc url, remove the trailing slash if present
	rpcURL = strings.TrimSuffix(rpcURL, "/")

	// default timeout is 10 seconds
	// maybe increase it later?
	if timeout == nil {
		timeout = new(time.Duration)
		*timeout = 10 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	return &RpcGnoland{
		rpcURL: rpcURL,
		client: &http.Client{
			Timeout:   *timeout,
			Transport: transport,
		},
	}, nil
}

// Only add methods that will be used by the indexer.
// Add future methods here.
const (
	Validators = "validators"
	Block      = "block"
	AbciQuery  = "abci_query"
	// might be useful for health check
	Health        = "health"
	Tx            = "tx"
	RequestCommit = "commit"
)

func (r *RpcGnoland) performRequest(method string, params map[string]any, result any) error {
	requestBody, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := r.client.Post(r.rpcURL, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error %s: %s", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Health method to get the health of the rpc client.
//
// Returns:
//   - error: if the call fails
func (r *RpcGnoland) Health() error {
	var response HealthResponse
	if err := r.performRequest(Health, nil, &response); err != nil {
		return err
	}
	if response.Error != nil {
		return fmt.Errorf("rpc error: %v, %s", response.Error.Code, response.Error.Message)
	}

	return nil
}

// GetValidators method to get validators from the rpc client.
//
// Parameters:
//   - height: the height of the block to get the validators for
//
// Returns:
//   - *ValidatorsResponse: the response from the rpc client
//   - error: if the call fails
func (r *RpcGnoland) GetValidators(height uint64) (*ValidatorsResponse, *RpcHeightError) {
	response := &ValidatorsResponse{}
	// convert the height to a string because the rpc client expects a string
	params := map[string]any{
		"height": strconv.FormatUint(height, 10),
	}
	if err := r.performRequest(Validators, params, response); err != nil {
		return nil, &RpcHeightError{
			Height:    height,
			HasHeight: true,
			Err:       err,
		}
	}
	if response.Error != nil {
		return nil, &RpcHeightError{
			Height:    height,
			HasHeight: true,
			Err:       fmt.Errorf("rpc error: %v, %s", response.Error.Code, response.Error.Message),
		}
	}
	return response, nil
}

// GetBlock method to get a block from the rpc client.
//
// Parameters:
//   - height: the height of the block to get
//
// Returns:
//   - *BlockResponse: the response from the rpc client
//   - error: if the call fails
func (r *RpcGnoland) GetBlock(height uint64) (*BlockResponse, *RpcHeightError) {
	response := &BlockResponse{}
	// convert the height to a string because the rpc client expects a string
	params := map[string]any{
		"height": strconv.FormatUint(height, 10),
	}
	if err := r.performRequest(Block, params, response); err != nil {
		return nil, &RpcHeightError{
			Height:    height,
			HasHeight: true,
			Err:       err,
		}
	}
	if response.Error != nil {
		return nil, &RpcHeightError{
			Height:    height,
			HasHeight: true,
			Err:       fmt.Errorf("rpc error: %v, %s", response.Error.Code, response.Error.Message),
		}
	}
	return response, nil
}

// This is method similar to GetBlock but it doesn't require a height
// Whole purpose of this method is to get the latest block height from the rpc client
// without having to query the block itself
func (r *RpcGnoland) GetLatestBlockHeight() (uint64, *RpcHeightError) {
	response := &BlockResponse{}
	if err := r.performRequest(Block, nil, response); err != nil {
		return 0, &RpcHeightError{
			Height:    0,
			HasHeight: true,
			Err:       err,
		}
	}
	if response.Error != nil {
		return 0, &RpcHeightError{
			Height:    0,
			HasHeight: true,
			Err:       fmt.Errorf("rpc error: %v, %s", response.Error.Code, response.Error.Message),
		}
	}
	height, err := response.GetHeight()
	if err != nil {
		return 0, &RpcHeightError{
			Height:    0,
			HasHeight: true,
			Err:       err,
		}
	}
	return height, nil
}

// GetTx method to get a tx from the rpc client.
//
// Parameters:
//   - txHash: the base64 encoded string of the tx to get
//
// Returns:
//   - *TxResponse: the response from the rpc client
//   - error: if the call fails
func (r *RpcGnoland) GetTx(txHash string) (*TxResponse, *RpcStringError) {
	response := &TxResponse{}
	params := map[string]any{
		"hash": txHash,
	}
	if err := r.performRequest(Tx, params, response); err != nil {
		return nil, &RpcStringError{
			Value:    txHash,
			HasValue: true,
			Err:      err,
		}
	}
	if response.Error != nil {
		return nil, &RpcStringError{
			Value:    txHash,
			HasValue: true,
			Err:      fmt.Errorf("rpc error: %v, %s", response.Error.Code, response.Error.Message),
		}
	}
	return response, nil
}

// GetAbciQuery method to get a abci query from the rpc client.
//
// This method is used to get any kind of data from the rpc client.
// This might not be used in the indexer but let's keep it here for now.
//
// Parameters:
//   - path: the path of the abci query
//   - data: the data of the abci query
//   - height: the height of the block to get the abci query for(optional, if not specified it will get the latest block)
//   - prove: whether to prove the abci query(optional, if not specified it will not prove the abci query, if true it will return a proof)
//
// Returns:
//   - any: the response from the rpc client, it can be a different type depending on the path and data
//   - error: if the call fails
func (r *RpcGnoland) GetAbciQuery(path string, data string, height *uint64, prove *bool) (any, error) {
	params := map[string]any{
		"path": path,
		"data": data,
	}
	if height != nil {
		params["height"] = fmt.Sprintf("%d", *height)
	}
	if prove != nil {
		params["prove"] = *prove
	}

	var response map[string]any
	if err := r.performRequest(AbciQuery, params, &response); err != nil {
		return nil, err
	}
	if err, ok := response["error"]; ok {
		return nil, fmt.Errorf("rpc error: %v", err)
	}

	return response["result"], nil
}

func (r *RpcGnoland) GetCommit(height uint64) (*CommitResponse, *RpcCommitError) {
	response := &CommitResponse{}
	params := map[string]any{
		"height": strconv.FormatUint(height, 10),
	}
	if err := r.performRequest(RequestCommit, params, response); err != nil {
		return nil, &RpcCommitError{
			Height:    height,
			HasHeight: true,
			Err:       err,
		}
	}
	if response.Error != nil {
		return nil, &RpcCommitError{
			Height:    height,
			HasHeight: true,
			Err:       fmt.Errorf("rpc error: %v, %s", response.Error.Code, response.Error.Message),
		}
	}
	return response, nil
}
