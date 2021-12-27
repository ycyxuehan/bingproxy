package proxy

import "net"

type ListenerMap map[string]net.Listener

func NewListenerMap()ListenerMap{
	return make(ListenerMap)
}