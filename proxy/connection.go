package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
)

type Protocol string

const (
	TCP = "TCP"
	UDP = "UDP"
)

type Connection struct {
	Proxy  string   `json:"proxy"`
	Client string   `json:"client"`
	Hash   string   `json:"hash"`
	Remote net.Conn `json:"-"`
	Local  net.Conn `json:"-"`
	ctx    context.Context
}

func NewConnection(ctx context.Context, hash string, client string,  remote net.Conn, local net.Conn) Connection {
	connection := Connection{
		Hash:   hash,
		ctx:    ctx,
		Remote: remote,
		Local:  local,
		Client: client,
	}
	return connection
}

func NewCommandConnection(ctx context.Context, data *Command) (*Connection, error) {
	if data == nil {
		return nil, fmt.Errorf(" connect data is nil")
	}
	connection := Connection{
		Hash: data.Name,
		ctx:  ctx,
	}
	log.Println("connect to local port", data.LocalAddr)
	localConn, err := net.Dial("tcp", data.LocalAddr)
	if err != nil {
		return nil, err
	}
	connection.Local = localConn
	log.Println("connect to remote port", data.RemoteAddr)
	remoteConn, err := net.Dial("tcp", data.RemoteAddr)
	if err != nil {
		return nil, err
	}
	connection.Remote = remoteConn
	return &connection, nil
}

func (c *Connection) Join(ctx context.Context, errChan chan error) {
	log.Println("join connection")
	go JoinConnection(ctx, c.Local, c.Remote, errChan)
	go JoinConnection(ctx, c.Remote, c.Local, errChan)
}

func (c *Connection) Release(errChan chan error) {
	err := c.Local.Close()
	SendError(errChan, err)
	err = c.Remote.Close()
	SendError(errChan, err)
}

type ConnectionMap map[string]*Connection

func NewConnectionMap() ConnectionMap {
	return make(ConnectionMap)
}

func (cm ConnectionMap) Get(hash string) (*Connection, error) {
	if conn, ok := cm[hash]; ok {
		return conn, nil
	}
	return nil, fmt.Errorf("not found")
}

func (cm ConnectionMap)Set(conn *Connection) (error) {
	if _, ok := cm[conn.Hash]; ok {
		return fmt.Errorf("conn exists")
	}
	cm[conn.Hash] = conn
	return nil
}
