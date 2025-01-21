package bridges

// BridgeProtocol represents supported bridge protocols
// type BridgeProtocol string

// const (
//     WBTC     BridgeProtocol = "WBTC"
//     RenBridge BridgeProtocol = "RenBridge"
//     Multichain BridgeProtocol = "Multichain"
// )

type BridgeProtocol int

const (
    WBTC BridgeProtocol = iota
    RenBridge
)
// ProtocolConfig stores bridge-specific configurations
type ProtocolConfig struct {
    ContractAddress string
    ABI  string
    Methods  map[string][4]byte
}

// Protocol configurations
var ProtocolConfigs = map[BridgeProtocol]ProtocolConfig{
    WBTC: {
        ContractAddress: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
        ABI: `[{"constant":false,"inputs":[{"name":"_btcAddr","type":"string"},{"name":"_value","type":"uint256"}],"name":"burn","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
        Methods: map[string][4]byte{
            "burn": {0x42, 0x96, 0x6c, 0x68},
        },
    },
    RenBridge: {
        ContractAddress: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
        ABI: `[{"inputs":[{"internalType":"bytes","name":"_msg","type":"bytes"},{"internalType":"bytes","name":"_sig","type":"bytes"},{"internalType":"bytes","name":"_btcAddr","type":"bytes"}],"name":"submit","outputs":[],"stateMutability":"nonpayable","type":"function"}]`,
        Methods: map[string][4]byte{
            "submit": {0x6a, 0x75, 0x5a, 0x5e},
        },
    },
}
