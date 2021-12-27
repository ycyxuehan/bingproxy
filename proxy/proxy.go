package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
)

type Proxy struct {
	Name            string `json:"name"`
	Client          string `json:"client"`
	InternalService string `json:"internalService"`
	ExternalPort    string `json:"ExternalPort"`
	Listener        net.Listener
}

func NewProxy(name string, service string, port string) (Proxy, error) {
	p := Proxy{
		Name:            name,
		InternalService: service,
		ExternalPort:    port,
	}
	return p, p.Check()
}

func (p *Proxy) Check() error {
	if p.Name == "" {
		return fmt.Errorf("must specify the name")
	}
	if p.InternalService == "" {
		return fmt.Errorf("must specify the internal service address")
	}
	if p.ExternalPort == "" {
		return fmt.Errorf("must specify the external port")
	}
	return nil
}

func (p *Proxy) Listen() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", p.ExternalPort))
	if err != nil {
		return err
	}
	p.Listener = lis
	return nil
}

func (p *Proxy) Accept(connChan chan Connection, errChan chan error) error {
	for {
		conn1, err := p.Listener.Accept()
		if err != nil {
			SendError(errChan, err)
			continue
		}
		connection := NewConnection(context.Background(), conn1.RemoteAddr().String(), p.Client, nil, conn1)
		log.Println(conn1.RemoteAddr().String(), p.InternalService)
		connection.Proxy = p.Name
		SendConn(connChan, connection) //发生连接，用于proxy
		//等待下一个连接
		conn2, err := p.Listener.Accept()
		if err != nil {
			SendError(errChan, err)
			continue
		}
		connection.Remote = conn2
		SendConn(connChan, connection) //发生连接，用于proxy
	}
}

type ProxyMap map[string]*Proxy

func NewProxyMap() ProxyMap {
	return make(ProxyMap)
}

func (sm ProxyMap) Get(name string) (*Proxy, error) {
	if s, ok := sm[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("cannot get proxy named %s, not found", name)
}
