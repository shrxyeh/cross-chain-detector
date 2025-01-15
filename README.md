
# Cross-Chain Transaction Monitor

Monitors cross-chain transactions for specified addresses across multiple blockchains.

## Prerequisites
- Go 1.19 or higher
- Access to blockchain node RPCs (Infura, BSC, # Cross-Chain Transaction Detector

A Go-based backend system for monitoring and detecting cross-chain transactions between Bitcoin and Ethereum networks, with THORChain DEX integration.

## Features

- Real-time monitoring of Bitcoin and Ethereum addresses
- Detection of cross-chain transactions through:
  - Bridge contracts (Wormhole, Polygon Bridge, Avalanche Bridge)
  - Wrapped tokens (WBTC)
  - THORChain swaps
- 30-second refresh interval for transaction checks
- Automatic transaction deduplication
- Detailed logging of cross-chain activities

## Prerequisites

- Go 1.19 or higher
- Access to blockchain API endpoints:
  - Alchemy API (Ethereum)
  - BlockCypher API (Bitcoin)
  - THORChain API

## Installation

1. Clone the repository:
```bash
git clone https://github.com/shrxyeh/cross-chain-detector.git
cd cross-chain-detector
```

2. Install dependencies:
```bash
go mod download
```

3. Create and configure environment variables:
```bash
cp .env.example .env
```

4. Update `.env` with your API keys and configuration:
```env
# Ethereum Configuration
ETH_API_KEY=your_ethereum_api_key
ETHEREUM_RPC=https://eth-mainnet.g.alchemy.com/v2/your_api_key

# Bitcoin Configuration
BTC_API_KEY=your_bitcoin_api_key
BITCOIN_RPC=https://api.blockcypher.com/v1/btc/main

# THORChain Configuration
THORCHAIN_API=https://thorchain.net

# Monitoring Configuration
MONITOR_ADDRESS=your_address_to_monitor
```

## Building and Running

1. Run the detector:
```bash
go run cmd/main.go
```

## Project Structure

```
cross-chain-detector/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── chains/
│   │   ├── bitcoin.go          # Bitcoin client implementation
│   │   └── ethereum.go         # Ethereum client implementation
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── monitor/
│   │   └── monitor.go          # Transaction monitoring logic
│   ├── thorchain/
│   │   └── client.go           # THORChain client
│   └── types/
│       └── types.go            # Common type definitions
├── .env                        # Environment configuration
├── go.mod                      # Go module file
└── README.md                   # This file
```

## Usage Example

1. Set up your environment variables in `.env`
2. Run the Project
```bash
go run cmd/main.go
```

3. The system will start monitoring the specified address and output detected cross-chain transactions:
```
CROSS-CHAIN TRANSACTION DETECTED ✓
----------------------------------------
Transaction Details:
- Hash: 0x1234...
- From: 0x742d...
- To: 0x3ee1...
- Value: 1.5 ETH

Cross-Chain Information:
- Source Chain: Ethereum
- Target Chain: Bitcoin
- Target Address: bc1q...
- Protocol: Wormhole
- Status: Pending
----------------------------------------
```

## Configuration Options

- `MONITOR_ADDRESS`: The address to monitor for cross-chain transactions
- `ETH_API_KEY`: Your Ethereum API key for Alchemy
- `BTC_API_KEY`: Your Bitcoin API key for BlockCypher
- `ETHEREUM_RPC`: Ethereum RPC endpoint URL
- `BITCOIN_RPC`: Bitcoin API endpoint URL
- `THORCHAIN_API`: THORChain API endpoint URL

## Error Handling

The system implements comprehensive error handling:
- API failures are logged and retried
- Invalid transactions are skipped
- Network issues are handled gracefully
- Graceful shutdown on system interruption

## Limitations

Current version limitations:
- Supports only Bitcoin and Ethereum networks
- Basic THORChain integration
- Limited number of supported bridge contracts
- Basic pattern detection algorithms



## Acknowledgments

- [Alchemy API](https://www.alchemy.com/) for Ethereum data
- [BlockCypher](https://www.blockcypher.com/) for Bitcoin data
- [THORChain](https://thorchain.org/) for DEX integrationetc.)

## Setup
1. Clone the repository
2. Update config/config.yaml with your RPC endpoints
3. Run `go mod tidy`
4. Run `go run cmd/main.go`
