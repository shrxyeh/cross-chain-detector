package types
import (
    "context" 
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

type CrossChainSwap struct {
    SourceChain     string
    TargetChain     string
    SourceAddress   string
    TargetAddress   string
    SwapID          string
    Status          string
}

type CrossChainInfo struct {
    SourceChain   string
    TargetChain   string
    TargetAddress string
    SwapID        string
    Status        string
    Protocol      string  
}

type Config struct {
    BitcoinRPC     string
    EthereumRPC    string
    ThorchainAPI   string
    MonitorAddress string
}