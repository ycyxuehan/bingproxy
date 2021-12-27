package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"

	"proxy.bing89.com/proxy"
)

const (
	DefaultChanLength = 5
)

type Server struct {
	connection net.Listener	//客户端连接
	clientMap proxy.ClientMap //客户端表，存储客户端的代理配置
	portMap map[string]string //端口表，速查客户端
	connectionChan chan proxy.Connection
	commandChan chan proxy.Command
	errChan chan error
	port string
	connectionPool proxy.ConnectionMap
}

func NewServer(port string, proxyConf string, errChan chan error)(*Server, error){
	svr := Server{
		portMap: make(map[string]string),
		connectionChan: make(chan proxy.Connection, DefaultChanLength),
		commandChan: make(chan proxy.Command, DefaultChanLength),
		connectionPool: proxy.NewConnectionMap(),
		port: port,
		errChan: errChan,
	}
	//
	err := proxy.KeepFileExists(proxyConf);
	if err != nil {
		return nil, err
	}

	clientMap, err := proxy.NewClientMapByFile(proxyConf)
	if err != nil {
		return nil, err
	}
	svr.clientMap = clientMap

	err = svr.CreateListener()
	return &svr, err
}

func (s *Server)CreateListener()error{
	address := fmt.Sprintf("0.0.0.0:%s", s.port)
	log.Println("listen ", address)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.connection = conn
	return nil
}

func (s *Server)Run(){
	go s.ConnectionProcess()
	for{
		tcpConn, err := s.connection.Accept()
		if err != nil {
			proxy.SendError(s.errChan, err)
			continue
		}
		log.Println(tcpConn.RemoteAddr().String())
		go s.ReadConnection(tcpConn)
	}
}

func (s *Server)ReadConnection(conn net.Conn)error{
	for {
		data := [256]byte{}
		reader := bufio.NewReader(conn)
		n, err :=reader.Read(data[:])
		if err != nil {
			proxy.SendError(s.errChan, err)
			continue
		}
		log.Println(string(data[:n]))
		cmd, err := proxy.NewBytesCommand(data[:n])
		if err != nil {
			proxy.SendError(s.errChan, err)
			continue
		}
		log.Println(cmd.String())
		switch cmd.Type {
		case proxy.KeepAlive:
			_, err = conn.Write(data[:n])
			if err != nil {
				proxy.SendError(s.errChan, err)
			}
		case proxy.Register:
			s.clientMap.AddClient(cmd.Name, conn)
			s.InitProxy(cmd.Name)
		}

	}
}

func (s *Server)DoProxy(p *proxy.Proxy)error{
	err := p.Listen()
	if err != nil {
		return err
	}
	go p.Accept(s.connectionChan, s.errChan)
	return nil
}

func (s *Server)ConnectionProcess()error{
	for{
		select{
		case conn := <- s.connectionChan:
			if conn.Remote != nil {	//是客户端的连接
				connection, err := s.connectionPool.Get(conn.Hash)
				if err != nil {
					proxy.SendError(s.errChan, err)
					continue
				}
				connection.Remote = conn.Remote
				connection.Join(context.Background(), s.errChan)
				continue
			}
			//发送command给客户端，通知需要连接
			log.Println("发送command给客户端，通知需要连接")
			log.Println("获取client")
			prx, err := s.clientMap.GetProxy(conn.Client, conn.Proxy)
			if err != nil {
				proxy.SendError(s.errChan, err)
				continue
			}
			command := proxy.NewCommand(conn.Hash, proxy.Connect, prx.ExternalPort, prx.InternalService)
			err = s.connectionPool.Set(&conn)
			if err != nil {
				proxy.SendError(s.errChan, err)
				continue
			}
			go s.SendCommand(conn.Client, command)
		}
	}
	
}

func (s *Server)SendCommand(client string, cmd proxy.Command)error{
	data, err := cmd.Bytes()
	if err != nil {
		return proxy.SendError(s.errChan, err)
	}
	clt, err := s.clientMap.Get(client)
	if err != nil {
		return proxy.SendError(s.errChan, err)
	}
	_, err = clt.Connection.Write(data)
	return proxy.SendError(s.errChan, err)
}

func (s *Server)InitProxy(cltName string)error{
	log.Println(s.clientMap, cltName)
	c, err := s.clientMap.Get(cltName)
	if err != nil {
		return proxy.SendError(s.errChan, err)
	}
	for _, p := range c.ProxyMap {
		go s.DoProxy(p)
	}
	return nil
}