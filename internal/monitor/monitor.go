package monitor

import (
    "context"
    "fmt"
    "log"
    "time"
    "strings"

    "github.com/shrxyeh/cross-chain-detector/internal/chains"
    "github.com/shrxyeh/cross-chain-detector/internal/thorchain"
    "github.com/shrxyeh/cross-chain-detector/internal/types"
)

type CrossChainMonitor struct {
    config       *types.Config
    bitcoin      types.TransactionChecker
    ethereum     types.TransactionChecker
    thorchain    *thorchain.Client
    processedTxs map[string]bool
}

// Known bridge contract addresses
var bridgeContracts = map[string]string{
    "0x3ee18B2214AFF97000D974cf647E7C347E8fa585": "Wormhole",
    "0x40ec5B33f54e0E8A33A975908C5BA1c14e5BbbDf": "Polygon Bridge",
    "0x88ad09518695c6c3712AC10a214bE5109a655671": "Avalanche Bridge",
}

// Known wrapped token contracts
var wrappedTokens = map[string]struct{
    WrappedToken  string
    OriginalChain string
}{
    "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599": { // WBTC
        WrappedToken:  "WBTC",
        OriginalChain: "Bitcoin",
    },
}

func NewCrossChainMonitor(cfg *types.Config) (*CrossChainMonitor, error) {
    eth, err := chains.NewEthereumClient(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("failed to create Ethereum client: %v", err)
    }

    thor, err := thorchain.NewClient(cfg.ThorchainAPI)
    if err != nil {
        return nil, fmt.Errorf("failed to create THORChain client: %v", err)
    }

    return &CrossChainMonitor{
        config:       cfg,
        ethereum:     eth,
        thorchain:    thor,
        processedTxs: make(map[string]bool),
    }, nil
}

func (m *CrossChainMonitor) CheckTransactions(ctx context.Context, address string) error {
    // Check if address is an Ethereum addressor not
    if !strings.HasPrefix(strings.ToLower(address), "0x") {
        return fmt.Errorf("unsupported address format: %s", address)
    }

    // Fetch Ethereum transactions
    ethTxs, err := m.ethereum.GetAddressTransactions(ctx, address)
    if err != nil {
        log.Printf("Warning: failed to get Ethereum transactions: %v", err)
        return err
    }

    // Process Ethereum transactions
    for _, tx := range ethTxs {
        if m.processedTxs[tx.Hash] {
            continue
        }
        
        // Mark as processed
        m.processedTxs[tx.Hash] = true

        // Check different cross-chain patterns
        if info := m.detectCrossChainPatterns(ctx, tx); info != nil {
            tx.CrossChainInfo = info
            m.handleCrossChainTransaction(tx)
        }
    }

    return nil
}

func (m *CrossChainMonitor) detectCrossChainPatterns(ctx context.Context, tx types.Transaction) *types.CrossChainInfo {
    // Pattern 1: Check THORChain swaps
    if info, err := m.thorchain.CheckCrossChainSwap(ctx, tx.Hash); err == nil && info != nil {
        return info
    }

    // Pattern 2: Check for known bridge contract addresses
    if info := m.checkBridgePatterns(tx); info != nil {
        return info
    }

    // Pattern 3: Check for token wrapping patterns
    if info := m.checkWrappingPatterns(tx); info != nil {
        return info
    }

    return nil
}

func (m *CrossChainMonitor) checkBridgePatterns(tx types.Transaction) *types.CrossChainInfo {
    // Check if transaction involves a known bridge contract
    if bridgeName, isBridge := bridgeContracts[strings.ToLower(tx.To)]; isBridge {
        return &types.CrossChainInfo{
            SourceChain:   tx.Chain,
            TargetChain:   "Unknown", // Would need additional logic to determine target chain
            TargetAddress: tx.From,   // Original sender typically receives on target chain
            SwapID:       tx.Hash,
            Status:       "Pending",  // Would need additional logic to track status
            Protocol:     bridgeName,
        }
    }
    return nil
}

func (m *CrossChainMonitor) checkWrappingPatterns(tx types.Transaction) *types.CrossChainInfo {
    // Check if transaction involves wrapped tokens
    if tokenInfo, isWrapped := wrappedTokens[strings.ToLower(tx.To)]; isWrapped {
        return &types.CrossChainInfo{
            SourceChain:   tx.Chain,
            TargetChain:   tokenInfo.OriginalChain,
            TargetAddress: tx.From, // Typically same address receives wrapped tokens
            SwapID:       tx.Hash,
            Status:       "Completed",
            Protocol:     fmt.Sprintf("%s Wrapping", tokenInfo.WrappedToken),
        }
    }
    return nil
}

func (m *CrossChainMonitor) handleCrossChainTransaction(tx types.Transaction) {
    // Determine if it's a cross-chain transaction
    isCrossChain := tx.CrossChainInfo != nil

    // Get value with proper formatting
    value := "Unknown"
    if tx.Value != "" {
        value = tx.Value
    }

    if isCrossChain {
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
            tx.Hash, tx.From, tx.To, value,
            tx.CrossChainInfo.SourceChain,
            tx.CrossChainInfo.TargetChain,
            tx.CrossChainInfo.TargetAddress,
            tx.CrossChainInfo.Protocol,
            tx.CrossChainInfo.Status)
    } else {
        log.Printf("\nRegular Transaction (No cross-chain activity)\n"+
            "----------------------------------------\n"+
            "- Hash: %s\n"+
            "- From: %s\n"+
            "- To: %s\n"+
            "- Value: %s\n"+
            "----------------------------------------\n",
            tx.Hash, tx.From, tx.To, value)
    }
}

func (m *CrossChainMonitor) MonitorAddress(ctx context.Context, address string) error {
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
