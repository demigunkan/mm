package vertex

import (
	"bytes"
	"encoding/json"
	"strings"
)

type Response[T any] struct {
	Id      int    `json:"id,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Channel string `json:"channel,omitempty"`
}

type Request[T any] struct {
	Id     int    `json:"id,omitempty"`
	Method Method `json:"method"`
	Stream T      `json:"stream,omitempty"`
}

func (r *Request[T]) Pack() []byte {
	b, _ := json.Marshal(r)
	return b
}

type OrderbookRequest struct {
	Type      string `json:"type"`
	ProductId int    `json:"product_id"`
	Depth     int    `json:"depth,omitempty"`
}

type OrderbookData struct {
	Type             string     `json:"type"`
	MinTimestamp     int        `json:"min_timestamp,string"`
	MaxTimestamp     int        `json:"max_timestamp,string"`
	LastMaxTimestamp int        `json:"last_max_timestamp,string"`
	ProductID        int        `json:"product_id"`
	Bids             [][]string `json:"bids"`
	Asks             [][]string `json:"asks"`
}

const (
	channelFieldName = "channel"
	dataFieldName    = "field"
)

type channelOnlyResponse struct {
	Channel string `json:"channel,omitempty"`
}

func GetChannel(data []byte, value *string) {
	if exitEarly := decodeChannel(data, value); exitEarly {
		r := &channelOnlyResponse{}
		if err := json.Unmarshal(data, &r); err != nil {
			return
		}

		*value = r.Channel
	}
}

func GetMainChannel(channel string) Channel {
	s := strings.Split(channel, ":")
	return Channel(s[0])
}

func decodeChannel(data []byte, value *string) bool {
	// Use a Decoder to decode the JSON data
	decoder := json.NewDecoder(bytes.NewReader(data))
	// decoder.
	for decoder.More() {
		// Read the next JSON value
		t, err := decoder.Token()
		if err != nil {
			return false
		}
		// If it's a key, check if it matches the field we're looking for
		if key, ok := t.(string); ok {
			if key == dataFieldName {
				return true
			}
			if key == channelFieldName {
				// Advance to the field's value
				if !decoder.More() {
					return false
				}
				// Decode the field's value
				if err := decoder.Decode(value); err != nil {
					return false
				}
				// Field found, return early
				return false
			}
		}
	}

	return false
}
