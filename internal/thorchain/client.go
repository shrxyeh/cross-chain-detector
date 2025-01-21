package thorchain

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/shrxyeh/cross-chain-detector/internal/types"
)

type Client struct {
    apiURL string
    client *http.Client
}

type THORChainSwap struct {
    InHash    string `json:"in_hash"`
    OutHash   string `json:"out_hash"`
    FromAddr  string `json:"from_address"`
    ToAddr    string `json:"to_address"`
    InChain   string `json:"in_chain"`
    OutChain  string `json:"out_chain"`
    Status    string `json:"status"`
    Type      string `json:"type"`
}

func NewClient(apiURL string) (*Client, error) {
    if apiURL == "" {
        return nil, fmt.Errorf("THORChain API URL is required")
    }

    return &Client{
        apiURL: apiURL,
        client: &http.Client{
            Timeout: time.Second * 10,
        },
    }, nil
}

func (c *Client) CheckCrossChainSwap(ctx context.Context, txHash string) (*types.CrossChainInfo, error) {
    //  API URL for swap details
    url := fmt.Sprintf("%s/swaps/%s", c.apiURL, txHash)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch swap details: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        // If it is not a cross-chain swap
        return nil, nil
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
    }

    var swap THORChainSwap
    if err := json.NewDecoder(resp.Body).Decode(&swap); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }

    // Only process BTC to ETH swaps
    if !isValidChainPair(swap.InChain, swap.OutChain) {
        return nil, nil
    }

    return &types.CrossChainInfo{
        SourceChain:   normalizeChainName(swap.InChain),
        TargetChain:   normalizeChainName(swap.OutChain),
        TargetAddress: swap.ToAddr,
        SwapID:       swap.OutHash,
        Status:       normalizeStatus(swap.Status),
        Protocol:     "THORChain",
    }, nil
}

// Helper functions for chain and status normalization
func normalizeChainName(chain string) string {
    chain = strings.ToUpper(chain)
    switch chain {
    case "BTC", "BITCOIN":
        return "BTC"
    case "ETH", "ETHEREUM":
        return "ETH"
    default:
        return chain
    }
}

func normalizeStatus(status string) string {
    status = strings.ToUpper(status)
    switch status {
    case "PENDING":
        return "Pending"
    case "COMPLETE", "SUCCESS":
        return "Completed"
    case "FAILED":
        return "Failed"
    default:
        return "Unknown"
    }
}

func isValidChainPair(inChain, outChain string) bool {
    inChain = normalizeChainName(inChain)
    outChain = normalizeChainName(outChain)
    
    validPairs := map[string]map[string]bool{
        "BTC": {"ETH": true},
        "ETH": {"BTC": true},
    }

    if pairs, ok := validPairs[inChain]; ok {
        return pairs[outChain]
    }
    return false
}
