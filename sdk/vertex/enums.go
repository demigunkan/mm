package vertex

type Method string

const (
	MethodPing        Method = "ping"
	MethodSubscribe   Method = "subscribe"
	MethodUnsubscribe Method = "unsubscribe"
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
	ChannelBookDepth Channel = "book_depth"
)
