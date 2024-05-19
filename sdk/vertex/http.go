package vertex

type MarketLiquidity struct {
	Status string `json:"status"`
	Data   struct {
		Bids      [][]string `json:"bids"`
		Asks      [][]string `json:"asks"`
		Timestamp int        `json:"timestamp,string"`
	} `json:"data"`
	RequestType string `json:"request_type"`
}
