package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
	"proxy.bing89.com/proxy"
)

type Client struct {
	Name string
	Server string
	ServerPort string
	connections proxy.ConnectionMap
	connection net.Conn
}

func NewClient(name string, server string, port string)(*Client, error){
	clt := &Client{
		Name: name,
		connections: proxy.NewConnectionMap(),
		Server: server,
		ServerPort: port,
		connection: nil,
	}
	log.Println("connect to server", server)
	err := clt.ConnectServer()
	if err != nil {
		return nil, err
	}
	log.Println("register client", name)
	err = clt.Register()
	return clt, err
}


func(c *Client)ConnectServer()error{
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.Server, c.ServerPort))
	if err != nil {
		return errors.WithMessage(err, "connect to server")
	}
	c.connection = conn
	return nil
}

func (c *Client)Register()error{
	command := proxy.Command{
		Name: c.Name,
		Type: proxy.Register,
	}
	data, err := command.Bytes()
	if err !=nil {
		return errors.WithMessage(err, "register client")
	}
	_, err = c.connection.Write(data)
	return errors.WithMessage(err, "register client")
}

func (c *Client)Run(errChan chan error)error{
	reader := bufio.NewReader(c.connection)
	commandChan := make(chan *proxy.Command)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int)
	//启动keepalive
	go c.KeepAlive(ctx, errChan, done)
	//启动command处理
	go c.DoCommand(ctx, errChan, commandChan)
	//等待command
	for {
		select {
		case <- done:
			cancel()
		default:
			data := [256]byte{}
			n, err := reader.Read(data[:])
			if err != nil {
				proxy.SendError(errChan, err)
				continue
			}
			cmd, err := proxy.NewBytesCommand(data[:n])
			if err != nil {
				proxy.SendError(errChan, err)
				continue
			}
			go func(){commandChan<- cmd}()
		}
	}
}

func (c *Client)ReleaseConnection(ctx context.Context, errChan chan error, command *proxy.Command)error{
	connection, err := c.connections.Get(command.Name)
	if err != nil {
		return proxy.SendError(errChan, err)
	}
	connection.Release(errChan)
	return nil
}

func (c *Client)ProxyFunc(ctx context.Context, errChan chan error, data *proxy.Command)error{
	data.RemoteAddr = fmt.Sprintf("%s:%s", c.Server, data.RemoteAddr) //把端口改成ip:port
	log.Println("command:", data.String())
	connection, err := proxy.NewCommandConnection(context.Background(), data)
	if err != nil {
		return proxy.SendError(errChan, err)
	}
	connection.Join(ctx,errChan)
	return nil
}

func (c *Client)DoCommand(ctx context.Context, errChan chan error, commandChan chan *proxy.Command){
	for {
		select {
		case proxyCommand := <- commandChan:
			log.Println("recive command:", proxyCommand.String())
			switch proxyCommand.Type {
			case proxy.Connect:
				go c.ProxyFunc(ctx, errChan, proxyCommand)
			case proxy.Release:
				go c.ReleaseConnection(ctx, errChan, proxyCommand)
			}
		case <- ctx.Done():
			return
		}
	}
}

func (c *Client)KeepAlive(ctx context.Context, errChan chan error, aliveChan chan int)error{
	for {
		time.Sleep(10 * time.Second)
		select {
		case <- ctx.Done():
			return nil
		default:
			cmd := proxy.NewCommand(c.Name, proxy.KeepAlive, "", "")
			data, err := cmd.Bytes()
			if err != nil {
				proxy.SendError(errChan, err)
				continue
			}
			_, err = c.connection.Write(data)
			if err != nil {
				proxy.SendError(errChan, err)
				aliveChan <- 0
				return err
			}
		}
	}
}