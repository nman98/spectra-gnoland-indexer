package rpcclient

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client/rate_limit"
)

// NewRateLimitedRpcClient creates a new rate-limited RPC client wrapper
//
// Parameters:
//   - rpcURL: the url of the rpc client
//   - timeout: the timeout for the rpc client (optional)
//   - requestsPerWindow: number of requests allowed per time window
//   - timeWindow: the time window for rate limiting (e.g., 1*time.Minute for 1 minute)
//
// Returns:
//   - *RateLimitedRpcClient: the rate-limited rpc client
//   - error: if the rpc client fails to connect
func NewRateLimitedRpcClient(
	rpcURL string,
	timeout *time.Duration,
	requestsPerWindow int,
	timeWindow time.Duration,
	userAgent *string,
) (*RateLimitedRpcClient, error) {
	// Create the underlying RPC client
	client, err := NewRpcClient(rpcURL, timeout, userAgent)
	if err != nil {
		return nil, err
	}

	// Create the rate limiter
	rateLimiter := rate_limit.NewChannelRateLimiter(requestsPerWindow, timeWindow)

	return &RateLimitedRpcClient{
		client:      client,
		rateLimiter: rateLimiter,
	}, nil
}

// Close properly shuts down the rate limiter
func (r *RateLimitedRpcClient) Close() {
	r.rateLimiter.Close()
}

// Health method with rate limiting
func (r *RateLimitedRpcClient) Health() error {
	r.rateLimiter.Wait() // This will block until a token is available
	return r.client.Health()
}

// GetValidators method with rate limiting
func (r *RateLimitedRpcClient) GetValidators(height uint64) (*ValidatorsResponse, *RpcHeightError) {
	r.rateLimiter.Wait()
	return r.client.GetValidators(height)
}

// GetBlock method with rate limiting
func (r *RateLimitedRpcClient) GetBlock(height uint64) (*BlockResponse, *RpcHeightError) {
	r.rateLimiter.Wait()
	return r.client.GetBlock(height)
}

// GetLatestBlockHeight method with rate limiting
func (r *RateLimitedRpcClient) GetLatestBlockHeight() (uint64, *RpcHeightError) {
	r.rateLimiter.Wait()
	return r.client.GetLatestBlockHeight()
}

// GetTx method with rate limiting
func (r *RateLimitedRpcClient) GetTx(txHash string) (*TxResponse, *RpcStringError) {
	r.rateLimiter.Wait()
	return r.client.GetTx(txHash)
}

// GetAbciQuery method with rate limiting
func (r *RateLimitedRpcClient) GetAbciQuery(path string, data string, height *uint64, prove *bool) (any, error) {
	r.rateLimiter.Wait()
	return r.client.GetAbciQuery(path, data, height, prove)
}

// GetCommit method with rate limiting
func (r *RateLimitedRpcClient) GetCommit(height uint64) (*CommitResponse, *RpcCommitError) {
	r.rateLimiter.Wait()
	return r.client.GetCommit(height)
}

// TryHealth - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryHealth() (error, bool) {
	if !r.rateLimiter.Allow() {
		return nil, false // rate limited
	}
	return r.client.Health(), true
}

// TryGetValidators - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryGetValidators(height uint64) (*ValidatorsResponse, *RpcHeightError, bool) {
	if !r.rateLimiter.Allow() {
		return nil, nil, false // rate limited
	}
	response, err := r.client.GetValidators(height)
	return response, err, true
}

// TryGetBlock - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryGetBlock(height uint64) (*BlockResponse, *RpcHeightError, bool) {
	if !r.rateLimiter.Allow() {
		return nil, nil, false // rate limited
	}
	response, err := r.client.GetBlock(height)
	return response, err, true
}

// TryGetLatestBlockHeight - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryGetLatestBlockHeight() (uint64, *RpcHeightError, bool) {
	if !r.rateLimiter.Allow() {
		return 0, nil, false // rate limited
	}
	response, err := r.client.GetLatestBlockHeight()
	return response, err, true
}

// TryGetTx - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryGetTx(txHash string) (*TxResponse, *RpcStringError, bool) {
	if !r.rateLimiter.Allow() {
		return nil, nil, false // rate limited
	}
	response, err := r.client.GetTx(txHash)
	return response, err, true
}

// TryGetAbciQuery - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryGetAbciQuery(path string, data string, height *uint64, prove *bool) (any, error, bool) {
	if !r.rateLimiter.Allow() {
		return nil, nil, false // rate limited
	}
	response, err := r.client.GetAbciQuery(path, data, height, prove)
	return response, err, true
}

// TryGetCommit - non-blocking version that returns false if rate limited
func (r *RateLimitedRpcClient) TryGetCommit(height uint64) (*CommitResponse, *RpcCommitError, bool) {
	if !r.rateLimiter.Allow() {
		return nil, nil, false // rate limited
	}
	response, err := r.client.GetCommit(height)
	return response, err, true
}

// GetRateLimiterStatus returns information about the current rate limiter status
func (r *RateLimitedRpcClient) GetRateLimiterStatus() rate_limit.ChannelRateLimiterStatus {
	return r.rateLimiter.GetStatus()
}
