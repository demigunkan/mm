package vertex

import (
	"github.com/demigunkan/mm/pkg/http"

	nethttp "net/http"
)

type Vertex struct {
	env env
}

func New(env env) *Vertex {
	return &Vertex{env: env}
}

func (v *Vertex) WS() string {
	return envs[v.env].ws
}

func (v *Vertex) HTTP() string {
	return envs[v.env].http
}

func SubscribeBookDepth(id int, productId int) *Request[*OrderbookRequest] {
	return &Request[*OrderbookRequest]{
		Id:     id,
		Method: MethodSubscribe,
		Stream: &OrderbookRequest{
			Type:      string(ChannelBookDepth),
			ProductId: productId,
		},
	}
}

// query?type=market_liquidity&product_id={product_id}&depth={depth}
func GetMarketLiquidity(productId int) (*MarketLiquidity, error) {
	http := http.New("https://gateway.prod.vertexprotocol.com/v1")
	res := &MarketLiquidity{}
	status, err := http.Request(nethttp.MethodGet,
		"/query?type=market_liquidity&product_id=4&depth=100",
		nil, res)
	if status != nethttp.StatusOK {
		return res, err
	}

	return res, err
}

// 	return &Request[*OrderbookRequest]{
// 		Id:        id,
// 		Method:    MethodSubscribe,
// 		ProductId: productId,
// 		Type:      "market_liquidity",
// 		Depth:     100,
// 	}
// }

// {
//   "type": "market_liquidity",
//   "product_id": 1,
//   "depth": 10
// }
