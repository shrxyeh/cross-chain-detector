package bridges

import (
    "bytes"  
    "context"
    "encoding/hex"
    "fmt"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/core/types"  
)

type BridgeDecoder struct {
    ethClient *ethclient.Client
    abis      map[BridgeProtocol]abi.ABI
}

func NewBridgeDecoder(ethRPC string) (*BridgeDecoder, error) {
    client, err := ethclient.Dial(ethRPC)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
    }

    decoder := &BridgeDecoder{
        ethClient: client,
        abis:      make(map[BridgeProtocol]abi.ABI),
    }

    // Initialize ABIs
    for protocol, config := range ProtocolConfigs {
        parsedABI, err := abi.JSON(strings.NewReader(config.ABI))
        if err != nil {
            return nil, fmt.Errorf("failed to parse ABI for %s: %v", protocol, err)
        }
        decoder.abis[protocol] = parsedABI
    }

    return decoder, nil
}

func (d *BridgeDecoder) DecodeTransaction(ctx context.Context, tx common.Hash, protocol BridgeProtocol) (string, error) {
    // Fetch transaction
    receipt, err := d.ethClient.TransactionReceipt(ctx, tx)
    if err != nil {
        return "", fmt.Errorf("failed to fetch transaction receipt: %v", err)
    }

    // Get transaction
    transaction, _, err := d.ethClient.TransactionByHash(ctx, tx)
    if err != nil {
        return "", fmt.Errorf("failed to fetch transaction: %v", err)
    }

    //Takes input data
    input := transaction.Data()
    
    // Decode based on protocol
    switch protocol {
    case WBTC:
        return d.decodeWBTCTransaction(input)
    case RenBridge:
        return d.decodeRenBridgeTransaction(input, receipt)
    default:
        return "", fmt.Errorf("unsupported protocol: %s", protocol)
    }
}

func (d *BridgeDecoder) decodeWBTCTransaction(input []byte) (string, error) {
    if len(input) < 4 {
        return "", fmt.Errorf("input too short")
    }

    methodSig := input[:4]
    expectedSig := ProtocolConfigs[WBTC].Methods["burn"] 
    if !bytes.Equal(methodSig, expectedSig[:]) { 
        return "", fmt.Errorf("not a WBTC burn transaction")
    }

    decoded, err := d.abis[WBTC].Methods["burn"].Inputs.Unpack(input[4:])
    if err != nil {
        return "", fmt.Errorf("failed to decode parameters: %v", err)
    }

    if len(decoded) < 1 {
        return "", fmt.Errorf("missing BTC address parameter")
    }

    btcAddr, ok := decoded[0].(string)
    if !ok {
        return "", fmt.Errorf("invalid BTC address format")
    }

    return btcAddr, nil
}

func (d *BridgeDecoder) decodeRenBridgeTransaction(input []byte, receipt *types.Receipt) (string, error) {
    for _, log := range receipt.Logs {
        if strings.EqualFold(log.Address.Hex(), ProtocolConfigs[RenBridge].ContractAddress) {
            if len(log.Data) >= 32 {
                btcAddrBytes := log.Data[12:32] 
                return hex.EncodeToString(btcAddrBytes), nil
            }
        }
    }
    
    return "", fmt.Errorf("BTC address not found in transaction logs")
}
