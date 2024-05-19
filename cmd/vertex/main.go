package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/demigunkan/mm/internal/types"
	"github.com/demigunkan/mm/pkg/orderbook"
	"github.com/demigunkan/mm/sdk/vertex"
	"github.com/lxzan/gws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	v := vertex.New(vertex.Mainnet)
	ws := &WebSocket{
		output: make(chan *vertex.OrderbookData, 10000),
	}

	socket, _, err := gws.NewClient(ws, &gws.ClientOption{
		Addr: v.WS(),
		PermessageDeflate: gws.PermessageDeflate{
			Enabled:               true,
			ServerContextTakeover: true,
			ClientContextTakeover: true,
		},
	})
	if err != nil {
		panic(err)
	}
	go socket.ReadLoop()

	payload := vertex.SubscribeBookDepth(1, 4)

	// fmt.Println("---", string(payload.Pack()))
	socket.WriteMessage(gws.OpcodeText, payload.Pack())

	liquidity, err := vertex.GetMarketLiquidity(4)
	if err != nil {
		panic(err)
	}

	fee := 0.0002
	ob := orderbook.New()

	for _, l := range liquidity.Data.Bids {
		price := parseFloat(l[0])
		netPrice := price * (1 + fee)
		ob.ModifyLevel(types.Side__BID, price, netPrice, parseFloat(l[1]))
	}
	for _, l := range liquidity.Data.Asks {
		price := parseFloat(l[0])
		netPrice := price * (1 - fee)
		ob.ModifyLevel(types.Side__ASK, price, netPrice, parseFloat(l[1]))
	}

	timestamp := liquidity.Data.Timestamp

	for {
		select {
		case response := <-ws.output:
			// fmt.Println(response.MinTimestamp, timestamp)
			if response.MinTimestamp > timestamp {
				for _, l := range response.Bids {
					price := parseFloat(l[0])
					netPrice := price * (1 + fee)
					ob.ModifyLevel(types.Side__BID, price, netPrice, parseFloat(l[1]))
				}
				for _, l := range response.Asks {
					price := parseFloat(l[0])
					netPrice := price * (1 - fee)
					ob.ModifyLevel(types.Side__ASK, price, netPrice, parseFloat(l[1]))
				}

				ob.Print()
			}

			// fmt.Printf("\r")
			// fmt.Print("latency ", time.Since(response.Tim).Nanoseconds(), "\n")
			//
		case <-ctx.Done():
			return
		}
	}
}

type WebSocket struct {
	output chan *vertex.OrderbookData
}

func (c *WebSocket) OnClose(socket *gws.Conn, err error) {
	fmt.Printf("onerror: err=%s\n", err.Error())
}

func (c *WebSocket) OnPong(socket *gws.Conn, payload []byte) {
}

func (c *WebSocket) OnOpen(socket *gws.Conn) {
}

func (c *WebSocket) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.WritePong(payload)
}

func (c *WebSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	// var channel string
	data := message.Bytes()
	// fmt.Println("---", string(data))
	// vertex.GetChannel(data, &channel)

	// switch vertex.GetMainChannel(channel) {
	// case vertex.ChannelBookDepth:
	if bytes.Contains(data, []byte("book_depth")) {
		response := &vertex.OrderbookData{}
		if err := json.Unmarshal(data, &response); err != nil {
			fmt.Printf("error: %s\n", err.Error())
		} else {
			c.output <- response
		}
	}

	// }
}

var pad = []rune("000000000000000000")

func parseFloat(s string) float64 {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("---", r, s)
		}
	}()

	if s == "0" {
		return 0
	}
	runes := []rune(s)
	length := len(runes)
	if length > 16 {
		length = length - 16
		runes = runes[:length]
	} else {
		length = 16 - length
		runes = append(pad[:length], runes...)
	}

	str := string(runes[:length-2]) + "." + string(runes[length-2:])

	val, _ := strconv.ParseFloat(str, 64)
	return val
}
