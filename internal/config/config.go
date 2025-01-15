package config

import (
    "fmt"
    "os"
    "github.com/joho/godotenv"
    "github.com/shrxyeh/cross-chain-detector/internal/types"
)
func LoadConfig() (*types.Config, error) {
    if err := godotenv.Load(); err != nil {
        fmt.Printf("Warning: .env file not found or error loading it: %v\n", err)
    }

    config := &types.Config{
        BitcoinRPC:     os.Getenv("BITCOIN_RPC"),
        EthereumRPC:    os.Getenv("ETHEREUM_RPC"),
        ThorchainAPI:   os.Getenv("THORCHAIN_API"),
        MonitorAddress: os.Getenv("MONITOR_ADDRESS"),
    }

    // Validate required configuration
    if config.BitcoinRPC == "" {
        return nil, fmt.Errorf("BITCOIN_RPC environment variable is required")
    }
    if config.EthereumRPC == "" {
        return nil, fmt.Errorf("ETHEREUM_RPC environment variable is required")
    }
    if config.ThorchainAPI == "" {
        return nil, fmt.Errorf("THORCHAIN_API environment variable is required")
    }

    return config, nil
}