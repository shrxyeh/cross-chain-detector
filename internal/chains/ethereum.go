package chains

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/shrxyeh/cross-chain-detector/internal/types"
)

type EthereumClient struct {
    apiURL string
    apiKey string
    client *http.Client
}

type AlchemyTransfer struct {
    From     string      `json:"from"`
    To       string      `json:"to"`
    Hash     string      `json:"hash"`
    Value    json.Number `json:"value"`
    Category string      `json:"category"`
}

type AlchemyResponse struct {
    JsonRPC string `json:"jsonrpc"`
    Result  struct {
        Transfers []AlchemyTransfer `json:"transfers"`
    } `json:"result"`
    Error *struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    } `json:"error"`
}

func NewEthereumClient(apiURL string) (*EthereumClient, error) {
    if apiURL == "" {
        return nil, fmt.Errorf("Ethereum API URL is required")
    }

    return &EthereumClient{
        apiURL: apiURL,
        apiKey: "UZN5XE7926FSWIBGSPT9BF6G6HYDJDHGBR",
        client: &http.Client{
            Timeout: time.Second * 10,
        },
    }, nil
}

func (c *EthereumClient) GetAddressTransactions(ctx context.Context, address string) ([]types.Transaction, error) {
    rpcRequest := struct {
        JsonRPC string        `json:"jsonrpc"`
        Method  string        `json:"method"`
        Params  []interface{} `json:"params"`
        ID      int          `json:"id"`
    }{
        JsonRPC: "2.0",
        Method:  "alchemy_getAssetTransfers",
        Params: []interface{}{
            map[string]interface{}{
                "fromAddress": address,
                "category":    []string{"external", "internal", "erc20", "erc721", "erc1155"},
                "maxCount":   "0x3e8", // hex for 1000
                "withMetadata": true,
            },
        },
        ID: 1,
    }

    reqBody, err := json.Marshal(rpcRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %v", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(reqBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch transactions: %v", err)
    }
    defer resp.Body.Close()

    var alchemyResp AlchemyResponse
    if err := json.NewDecoder(resp.Body).Decode(&alchemyResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }

    if alchemyResp.Error != nil {
        return nil, fmt.Errorf("API error: %s", alchemyResp.Error.Message)
    }

    var transactions []types.Transaction
    for _, transfer := range alchemyResp.Result.Transfers {
        transactions = append(transactions, types.Transaction{
            Chain:     "ETH",
            Hash:      transfer.Hash,
            From:      transfer.From,
            To:        transfer.To,
            Value:     transfer.Value.String(),
            Timestamp: time.Now().Unix(), 
        })
    }

    return transactions, nil
}