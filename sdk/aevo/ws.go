package aevo

import (
	"bytes"
	"strings"

	"github.com/goccy/go-json"
)

type Response[T any] struct {
	Id      int    `json:"id,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Channel string `json:"channel,omitempty"`
}

type Request[T any] struct {
	Id   int `json:"id,omitempty"`
	Op   Op  `json:"op"`
	Data T   `json:"data,omitempty"`
}

func (r *Request[T]) Pack() []byte {
	b, _ := json.Marshal(r)
	return b
}

type OrderbookData struct {
	Type           Type       `json:"type"`
	InstrumentID   uint32     `json:"instrument_id,string"`
	InstrumentName string     `json:"instrument_name"`
	InstrumentType string     `json:"instrument_type"`
	Bids           [][]string `json:"bids"`
	Asks           [][]string `json:"asks"`
	LastUpdated    int64      `json:"last_updated,string"`
	Checksum       string     `json:"checksum"`
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
