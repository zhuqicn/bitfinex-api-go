package bitfinex

import (
	"encoding/json"
	"fmt"
)

type eventType struct {
	Event string `json:"event"`
}

type InfoEvent struct {
	Version int `json:"version"`
	Code    int `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
}

type AuthEvent struct {
	Event   string       `json:"event"`
	Status  string       `json:"status"`
	ChanID  int64        `json:"chanId,omitempty"`
	UserID  int64        `json:"userId,omitempty"`
	SubID   string       `json:"subId"`
	AuthID  string       `json:"auth_id,omitempty"`
	Message string       `json:"msg,omitempty"`
	Caps    Capabilities `json:"caps"`
}

type Capability struct {
	Read  int `json:"read"`
	Write int `json:"write"`
}

type Capabilities struct {
	Orders    Capability `json:"orders"`
	Account   Capability `json:"account"`
	Funding   Capability `json:"funding"`
	History   Capability `json:"history"`
	Wallets   Capability `json:"wallets"`
	Withdraw  Capability `json:"withdraw"`
	Positions Capability `json:"positions"`
}

type ErrorEvent struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type UnsubscribeEvent struct {
	Status string `json:"status"`
	ChanID int64  `json:"chanId"`
}

type SubscribeEvent struct {
	SubID     string `json:"subId"`
	Channel   string `json:"channel"`
	ChanID    int64  `json:"chanId"`
	Symbol    string `json:"symbol"`
	Precision string `json:"prec,omitempty"`
	Frequency string `json:"freq,omitempty"`
	Key       string `json:"key,omitempty"`
	Len       string `json:"len,omitempty"`
	Pair      string `json:"pair"`
}

type ConfEvent struct {
	Status string `json:"status"`
}

// onEvent handles all the event messages and connects SubID and ChannelID.
func (b *bfxWebsocket) onEvent(msg []byte) (interface{}, error) {
	event := &eventType{}
	err := json.Unmarshal(msg, event)
	if err != nil {
		return nil, err
	}

	var e interface{}
	switch event.Event {
	case "info":
		infoEvent := InfoEvent{}
		err = json.Unmarshal(msg, &infoEvent)
		return infoEvent, err
	case "auth":
		// TODO: should the lib itself keep track of the authentication
		// 			 status?
		a := AuthEvent{}
		err = json.Unmarshal(msg, &a)
		if err != nil {
			return nil, err
		}

		b.subMu.Lock()
		if _, ok := b.privSubIDs[a.SubID]; ok {
			b.privChanIDs[a.ChanID] = struct{}{}
			delete(b.privSubIDs, a.SubID)
		}
		b.subMu.Unlock()
		return a, nil
	case "subscribed":
		s := SubscribeEvent{}
		err = json.Unmarshal(msg, &s)
		if err != nil {
			return nil, err
		}

		b.subMu.Lock()
		if info, ok := b.pubSubIDs[s.SubID]; ok {
			b.pubChanIDs[s.ChanID] = info.req
			b.handlersMu.Lock()
			b.publicHandlers[s.ChanID] = info.h
			b.handlersMu.Unlock()
			delete(b.pubSubIDs, s.SubID)
		}
		b.subMu.Unlock()
		return s, nil
	case "unsubscribed":
		e = UnsubscribeEvent{}
	case "error":
		e = ErrorEvent{}
	case "conf":
		confEvent := ConfEvent{}
		err = json.Unmarshal(msg, &confEvent)
		return confEvent, err
	default:
		return nil, fmt.Errorf("unknown event: %s", msg) // TODO: or just log?
	}

	err = json.Unmarshal(msg, &e)
	return e, err
}
