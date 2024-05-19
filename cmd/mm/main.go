package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/demigunkan/mm/internal/interfaces"
	"github.com/demigunkan/mm/internal/types"
	"github.com/demigunkan/mm/pkg/orderbook"
	"github.com/demigunkan/mm/sdk/aevo"
	"github.com/lxzan/gws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := aevo.New(aevo.Testnet)
	ws := &WebSocket{
		output: make(chan *aevo.Response[*aevo.OrderbookData], 10000),
	}

	socket, _, err := gws.NewClient(ws, &gws.ClientOption{
		Addr: a.WS(),
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

	payload := aevo.SubscribeRequest(1, []aevo.Channel{aevo.ChannelOrderbook.WithArg("ETH-PERP")})

	socket.WriteMessage(gws.OpcodeText, payload.Pack())

	var ob interfaces.Orderbook

	fee := 0.0008
	for {
		select {
		case response := <-ws.output:
			if response.Data.Type == "snapshot" {
				ob = orderbook.New()
			}

			for _, l := range response.Data.Bids {
				price := parseFloat(l[0])
				netPrice := price * (1 - fee)
				ob.ModifyLevel(types.Side__BID, price, netPrice, parseFloat(l[1]))
			}
			for _, l := range response.Data.Asks {
				price := parseFloat(l[0])
				netPrice := price * (1 + fee)
				ob.ModifyLevel(types.Side__ASK, price, netPrice, parseFloat(l[1]))
			}

			// fmt.Printf("\r")
			// fmt.Print("latency ", time.Since(response.Tim).Nanoseconds(), "\n")
			ob.Print()
		case <-ctx.Done():
			return
		}
	}
}

type WebSocket struct {
	output chan *aevo.Response[*aevo.OrderbookData]
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

	var channel string
	data := message.Bytes()
	aevo.GetChannel(data, &channel)

	switch aevo.GetMainChannel(channel) {
	case aevo.ChannelOrderbook:
		response := &aevo.Response[*aevo.OrderbookData]{}
		if err := json.Unmarshal(data, &response); err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
		c.output <- response
	}
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
