package vertex

type env uint

const (
	Testnet env = iota
	Mainnet
)

type envConfig struct {
	http string
	ws   string
}

var envs = map[env]envConfig{
	Mainnet: {
		http: "https://api.aevo.xyz",
		ws:   "wss://gateway.prod.vertexprotocol.com/v1/subscribe",
	},
	Testnet: {
		http: "https://api-testnet.aevo.xyz",
		ws:   "wss://ws-testnet.aevo.xyz",
	},
}
