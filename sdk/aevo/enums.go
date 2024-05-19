package aevo

type Op string

const (
	OpPing        Op = "ping"
	OpSubscribe   Op = "subscribe"
	OpUnsubscribe Op = "unsubscribe"
)

type Type string

const (
	TypeSnapshot Type = "snapshot"
	TypeUpdate   Type = "update"
)

type Channel string

const channelSeparator = ":"

func (c Channel) WithArg(p string) Channel {
	return Channel(string(c) + channelSeparator + p)
}

const (
	ChannelOrderbook Channel = "orderbook"
)
