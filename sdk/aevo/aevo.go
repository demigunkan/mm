package aevo

import "github.com/ethereum/go-ethereum/signer/core/apitypes"

type Aevo struct {
	env env
}

func New(env env) *Aevo {
	return &Aevo{env: env}
}

func (a *Aevo) WS() string {
	return envs[a.env].ws
}

func (a *Aevo) HTTP() string {
	return envs[a.env].http
}

func (a *Aevo) Domain() apitypes.TypedDataDomain {
	return envs[a.env].domain
}

func SubscribeRequest(id int, chs []Channel) *Request[[]Channel] {
	return &Request[[]Channel]{
		Id:   id,
		Op:   OpSubscribe,
		Data: chs,
	}
}
