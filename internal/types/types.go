package types

import (
    "context"
    "github.com/shrxyeh/cross-chain-detector/internal/bridges"
)

type TransactionChecker interface {
    GetAddressTransactions(ctx context.Context, address string) ([]Transaction, error)
}

type Transaction struct {
    Chain           string
    Hash            string
    From            string
    To              string
    Value           string
    Timestamp       int64
    CrossChainInfo  *CrossChainInfo
}

type CrossChainInfo struct {
    SourceChain   string
    TargetChain   string
    TargetAddress string
    SwapID        string
    Status        string
    Protocol      string
}

type CrossChainPattern struct {
    SourceChain      string
    TargetChain      string
    Protocol         string
    DestinationCheck func(string) bool
}

type Config struct {
    BitcoinRPC     string
    EthereumRPC    string
    MonitorAddress string
}

type BridgeConfig struct {
    Name         string
    SourceChain  string
    TargetChain  string
    BridgeType   bridges.BridgeProtocol
    RequiresLogs bool
}
