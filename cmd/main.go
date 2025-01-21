package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/shrxyeh/cross-chain-detector/internal/config"
    "github.com/shrxyeh/cross-chain-detector/internal/monitor"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create a new monitor
    mon, err := monitor.NewCrossChainMonitor(cfg)
    if err != nil {
        log.Fatalf("Failed to create monitor: %v", err)
    }

    // Create context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown gracefully
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Get the address to monitor
    address := os.Getenv("MONITOR_ADDRESS")
    if address == "" {
        log.Fatal("MONITOR_ADDRESS environment variable not set")
    }

    // Start monitoring in a goroutine
    go func() {
        if err := mon.MonitorAddress(ctx, address); err != nil && err != context.Canceled {
            log.Printf("Monitoring stopped with error: %v", err)
        }
    }()

    // Wait for shutdown signal
    <-sigChan
    log.Println("Shutting down...")
    cancel()

    // Allow some time for cleanup
    time.Sleep(2 * time.Second)
    log.Println("Shutdown complete")
}
