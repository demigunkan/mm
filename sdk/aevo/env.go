package aevo

import (
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type env uint

const (
	Testnet env = iota
	Mainnet
)

type envConfig struct {
	http   string
	ws     string
	domain apitypes.TypedDataDomain
}

var envs = map[env]envConfig{
	Mainnet: {
		http: "https://api.aevo.xyz",
		ws:   "wss://ws.aevo.xyz",
		domain: apitypes.TypedDataDomain{
			Name:    "Aevo Mainnet",
			Version: "1",
			ChainId: math.NewHexOrDecimal256(1),
		},
	},
	Testnet: {
		http: "https://api-testnet.aevo.xyz",
		ws:   "wss://ws-testnet.aevo.xyz",
		domain: apitypes.TypedDataDomain{
			Name:    "Aevo Testnet",
			Version: "1",
			ChainId: math.NewHexOrDecimal256(11155111),
		},
	},
}
