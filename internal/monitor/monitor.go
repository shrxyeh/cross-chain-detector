package monitor

import (
    "context"
    "log"
    "time"

    "github.com/shrxyeh/cross-chain-detector/internal/thorchain"
    "github.com/shrxyeh/cross-chain-detector/internal/types"
)

type CrossChainMonitor struct {
    thorchainClient *thorchain.Client
    processedTxs   map[string]bool
}

func NewCrossChainMonitor(cfg *types.Config) (*CrossChainMonitor, error) {
    thor, err := thorchain.NewClient(cfg.THORChainAPIURL)
    if err != nil {
        return nil, err
    }

    return &CrossChainMonitor{
        thorchainClient: thor,
        processedTxs:    make(map[string]bool),
    }, nil
}

func (m *CrossChainMonitor) CheckTransactions(ctx context.Context, address string) error {
    // Check for cross-chain swaps via THORChain
    info, err := m.thorchainClient.CheckCrossChainSwap(ctx, address)
    if err != nil {
        return err
    }

    if info != nil {
        m.handleCrossChainTransaction(info)
    }

    return nil
}

func (m *CrossChainMonitor) handleCrossChainTransaction(info *types.CrossChainInfo) {
    log.Printf("\nCROSS-CHAIN TRANSACTION DETECTED ✓\n"+
        "----------------------------------------\n"+
        "Cross-Chain Information:\n"+
        "- Source Chain: %s\n"+
        "- Target Chain: %s\n"+
        "- Target Address: %s\n"+
        "- Swap ID: %s\n"+
        "- Status: %s\n"+
        "- Protocol: %s\n"+
        "----------------------------------------\n",
        info.SourceChain,
        info.TargetChain,
        info.TargetAddress,
        info.SwapID,
        info.Status,
        info.Protocol)
}

func (m *CrossChainMonitor) MonitorAddress(ctx context.Context, address string) error {
    log.Printf("Starting to monitor address: %s\n", address)

    if err := m.CheckTransactions(ctx, address); err != nil {
        log.Printf("Error in initial transaction check: %v", err)
    }

    // Create ticker for periodic checking
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := m.CheckTransactions(ctx, address); err != nil {
                log.Printf("Error checking transactions: %v", err)
                continue
            }
        }
    }
}
