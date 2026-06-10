package config_test

import (
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	"github.com/stretchr/testify/assert"
)

func TestNormalLoadConfig(t *testing.T) {
	conf, err := config.LoadConfig("testdata/test.yml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	assert.Equal(t, "https://gnoland-testnet-rpc.cogwheel.zone", conf.RpcUrl)
	assert.Equal(t, "gnoland-indexer", *conf.UserAgent)
	assert.Equal(t, 100, conf.PoolMaxConns)
	assert.Equal(t, 10, conf.PoolMinConns)
	assert.Equal(t, time.Second*60, conf.PoolMaxConnLifetime)
	assert.Equal(t, time.Second*60, conf.PoolMaxConnIdleTime)
	assert.Equal(t, time.Second*60, conf.PoolHealthCheckPeriod)
	assert.Equal(t, time.Second*60, conf.PoolMaxConnLifetimeJitter)
	assert.Equal(t, uint64(100), conf.MaxBlockChunkSize)
	assert.Equal(t, uint64(100), conf.MaxTransactionChunkSize)
	assert.Equal(t, "gnoland", conf.ChainName)

}

func TestErrorLoadConfig(t *testing.T) {
	// this test should fail because the config is invalid
	_, err := config.LoadConfig("testdata/test2.yml")
	assert.Error(t, err)

	_, err = config.LoadConfig("testdata/test3.yml")
	assert.Error(t, err)
}
