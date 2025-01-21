package monitor

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shrxyeh/cross-chain-detector/internal/chains"
    "github.com/shrxyeh/cross-chain-detector/internal/bridges"
    "github.com/shrxyeh/cross-chain-detector/internal/types"
)

type CrossChainMonitor struct {
    config       *types.Config
    bitcoin      types.TransactionChecker
    ethereum     types.TransactionChecker
    bridgeDecoder *bridges.BridgeDecoder
    processedTxs map[string]bool
}

// Known cross-chain bridge addresses and their configurations
var crossChainBridges = map[string]types.BridgeConfig{
    // WBTC Bridge
    "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599": {
        Name:         "WBTC",
        SourceChain:  "ETH",
        TargetChain:  "BTC",
        BridgeType:   bridges.WBTC,
        RequiresLogs: true,
    },
    // RenBridge
    "0x3ee18B2214AFF97000D974cf647E7C347E8fa585": {
        Name:         "RenBridge",
        SourceChain:  "ETH",
        TargetChain:  "BTC",
        BridgeType:   bridges.RenBridge,
        RequiresLogs: true,
    },
    // Bitcoin to Ethereum bridges
    "1FcXr8tDtXxQvuoXqC4sj5uQh5NNvZvfXu": {
        Name:         "RenBridge-BTC",
        SourceChain:  "BTC",
        TargetChain:  "ETH",
        BridgeType:   bridges.RenBridge,
        RequiresLogs: false,
    },
}

func NewCrossChainMonitor(cfg *types.Config) (*CrossChainMonitor, error) {
    btc, err := chains.NewBitcoinClient(cfg.BitcoinRPC)
    if err != nil {
        return nil, fmt.Errorf("failed to create Bitcoin client: %v", err)
    }

    eth, err := chains.NewEthereumClient(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("failed to create Ethereum client: %v", err)
    }

    decoder, err := bridges.NewBridgeDecoder(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("failed to create bridge decoder: %v", err)
    }

    return &CrossChainMonitor{
        config:       cfg,
        bitcoin:      btc,
        ethereum:     eth,
        bridgeDecoder: decoder,
        processedTxs: make(map[string]bool),
    }, nil
}

func (m *CrossChainMonitor) CheckTransactions(ctx context.Context, address string) error {
    var transactions []types.Transaction
    var err error

    // Determine chain type based on address format
    if isValidBitcoinAddress(address) {
        transactions, err = m.bitcoin.GetAddressTransactions(ctx, address)
        if err != nil {
            return fmt.Errorf("failed to get Bitcoin transactions: %v", err)
        }
    } else if isValidEthereumAddress(address) {
        transactions, err = m.ethereum.GetAddressTransactions(ctx, address)
        if err != nil {
            return fmt.Errorf("failed to get Ethereum transactions: %v", err)
        }
    } else {
        return fmt.Errorf("unsupported address format: %s", address)
    }

    for _, tx := range transactions {
        if m.processedTxs[tx.Hash] {
            continue
        }
        
        m.processedTxs[tx.Hash] = true

        // Detect and handle cross-chain transactions
        if info := m.detectCrossChainTransaction(ctx, tx); info != nil {
            tx.CrossChainInfo = info
            m.handleCrossChainTransaction(tx)
        }
    }

    return nil
}

func (m *CrossChainMonitor) detectCrossChainTransaction(ctx context.Context, tx types.Transaction) *types.CrossChainInfo {
    // Check if the transaction involves a known bridge
    if bridgeConfig, isBridge := crossChainBridges[tx.To]; isBridge {
        if tx.Chain == bridgeConfig.SourceChain {
            targetAddr, err := m.deriveTargetAddress(ctx, tx, bridgeConfig)
            if err != nil {
                log.Printf("Failed to derive target address: %v", err)
                targetAddr = "Unknown"
            }

            return &types.CrossChainInfo{
                SourceChain:   tx.Chain,
                TargetChain:   bridgeConfig.TargetChain,
                TargetAddress: targetAddr,
                SwapID:       tx.Hash,
                Status:       determineTransactionStatus(tx),
                Protocol:     bridgeConfig.Name,
            }
        }
    }

    return nil
}

func (m *CrossChainMonitor) deriveTargetAddress(ctx context.Context, tx types.Transaction, bridge types.BridgeConfig) (string, error) {
    if bridge.RequiresLogs {
        // For Ethereum transactions that require log analysis
        return m.bridgeDecoder.DecodeTransaction(ctx, common.HexToHash(tx.Hash), bridge.BridgeType)
    }

    // For Bitcoin transactions or simple transfers
    if tx.Chain == "BTC" {
        return m.extractEthAddressFromBTC(tx)
    }

    return "", fmt.Errorf("unsupported bridge configuration")
}

func (m *CrossChainMonitor) extractEthAddressFromBTC(tx types.Transaction) (string, error) {
    // This is a simplified implementation for demonstration purposes
    return "Unknown", fmt.Errorf("BTC to ETH address extraction not implemented")
}

func (m *CrossChainMonitor) handleCrossChainTransaction(tx types.Transaction) {
    log.Printf("\nCROSS-CHAIN TRANSACTION DETECTED ✓\n"+
        "----------------------------------------\n"+
        "Transaction Details:\n"+
        "- Hash: %s\n"+
        "- From: %s\n"+
        "- To: %s\n"+
        "- Value: %s\n"+
        "\nCross-Chain Information:\n"+
        "- Source Chain: %s\n"+
        "- Target Chain: %s\n"+
        "- Target Address: %s\n"+
        "- Protocol: %s\n"+
        "- Status: %s\n"+
        "----------------------------------------\n",
        tx.Hash, tx.From, tx.To, tx.Value,
        tx.CrossChainInfo.SourceChain,
        tx.CrossChainInfo.TargetChain,
        tx.CrossChainInfo.TargetAddress,
        tx.CrossChainInfo.Protocol,
        tx.CrossChainInfo.Status)
}

func (m *CrossChainMonitor) MonitorAddress(ctx context.Context, address string) error {
    log.Printf("Starting to monitor address: %s\n", address)
    
    // Initial check
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

// Helper functions
func isValidBitcoinAddress(address string) bool {
    // Basic validation for Bitcoin addresses
    if strings.HasPrefix(address, "0x") {
        return false
    }
    return len(address) >= 26 && len(address) <= 35
}

func isValidEthereumAddress(address string) bool {
    // Basic validation for Ethereum addresses
    return strings.HasPrefix(strings.ToLower(address), "0x") && len(address) == 42
}

func determineTransactionStatus(tx types.Transaction) string {
    // This could be developed further to check confirmation counts, bridge contract states, etc.
    return "Pending"
}

func bytesToAddress(b []byte) string {
    return common.BytesToAddress(b).Hex()
}
