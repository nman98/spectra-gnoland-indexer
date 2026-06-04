package rpcclient

import (
	"net/http"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client/rate_limit"
)

// RpcGnoland is the struct for the rpc client
type RpcGnoland struct {
	rpcURL string
	client *http.Client
}

// JsonRpcError is part of the struct for the rpc client
type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type HealthResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  any           `json:"result"`
}

// Client is the interface for the rpc client
type Client interface {
	Health() error
	GetValidators(height uint64) (*ValidatorsResponse, *RpcHeightError)
	GetBlock(height uint64) (*BlockResponse, *RpcHeightError)
	GetLatestBlockHeight() (uint64, *RpcHeightError)
	GetTx(txHash string) (*TxResponse, *RpcStringError)
	GetAbciQuery(path string, data string, height *uint64, prove *bool) (any, error)
	GetCommit(height uint64) (*CommitResponse, *RpcCommitError)
}

type RateLimiter interface {
	Allow() bool
	Wait()
	Close()
	GetStatus() rate_limit.ChannelRateLimiterStatus
}

// RateLimitedRpcClient wraps the original RPC client with rate limiting
//
// The struct contains the client and the rate limiter
// The client is the original RPC client
// The rate limiter is the rate limiter for the client
type RateLimitedRpcClient struct {
	client      Client
	rateLimiter RateLimiter
}
