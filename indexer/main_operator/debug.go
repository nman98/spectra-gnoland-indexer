//go:build devmode

package mainoperator

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func InitDebug(ctx context.Context) {
    l.Info().Msg("Debug mode started")
    go printMemStats(ctx, 5*time.Second)
    go http.ListenAndServe(":6060", nil)
    <-ctx.Done()
    l.Info().Msg("Debug mode stopped")
}

func printMemStats(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        defer ticker.Stop()
        var m runtime.MemStats
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                runtime.ReadMemStats(&m)
                fmt.Printf("[mem] alloc=%vMB sys=%vMB gc_cycles=%v\n",
                    m.Alloc/1024/1024,
                    m.Sys/1024/1024,
                    m.NumGC,
                )
            }
        }
    }()
}