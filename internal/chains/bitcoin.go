package chains

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/shrxyeh/cross-chain-detector/internal/types"
)

type BitcoinClient struct {
    apiURL  string
    apiKey  string
    client  *http.Client
}

type BlockCypherTx struct {
    Hash      string    `json:"hash"`
    Addresses []string  `json:"addresses"`
    Total     int64     `json:"total"`
    Confirmed time.Time `json:"confirmed"`
    Inputs    []struct {
        Addresses []string `json:"addresses"`
    } `json:"inputs"`
    Outputs []struct {
        Addresses []string `json:"addresses"`
    } `json:"outputs"`
}

type BlockCypherResponse struct {
    Address       string         `json:"address"`
    TotalReceived int64         `json:"total_received"`
    TotalSent     int64         `json:"total_sent"`
    Balance       int64         `json:"balance"`
    TXs          []BlockCypherTx `json:"txs"`
}

func NewBitcoinClient(apiURL string) (*BitcoinClient, error) {
    if apiURL == "" {
        return nil, fmt.Errorf("Bitcoin API URL is required")
    }
    
    return &BitcoinClient{
        apiURL: apiURL,
        apiKey: "8be8243b67f84057abd20244dd535421",
        client: &http.Client{
            Timeout: time.Second * 10,
        },
    }, nil
}

func (c *BitcoinClient) GetAddressTransactions(ctx context.Context, address string) ([]types.Transaction, error) {
    url := fmt.Sprintf("%s/addrs/%s?limit=50", c.apiURL, address)
    if c.apiKey != "" {
        url += "&token=" + c.apiKey
    }

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch transactions: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
    }

    var blockCypherResp BlockCypherResponse
    if err := json.NewDecoder(resp.Body).Decode(&blockCypherResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }

    var transactions []types.Transaction
    for _, tx := range blockCypherResp.TXs {
        from := ""
        if len(tx.Inputs) > 0 && len(tx.Inputs[0].Addresses) > 0 {
            from = tx.Inputs[0].Addresses[0]
        }

        to := ""
        if len(tx.Outputs) > 0 && len(tx.Outputs[0].Addresses) > 0 {
            to = tx.Outputs[0].Addresses[0]
        }

        transactions = append(transactions, types.Transaction{
            Chain:     "BTC",
            Hash:      tx.Hash,
            From:      from,
            To:        to,
            Value:     fmt.Sprintf("%d", tx.Total),
            Timestamp: tx.Confirmed.Unix(),
        })
    }

    return transactions, nil
}