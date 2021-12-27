package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)


type Client struct {
	Name string	`json:"name"`
	Address string	`json:"address"`
	Connection net.Conn
	ProxyMap ProxyMap	`json:"proxies"`
}

func NewClient(name string)Client{
	return Client{
		Name:  name,
		ProxyMap: NewProxyMap(),
	}
}

type ClientMap map[string]*Client

func NewClientMap()ClientMap{
	return make(ClientMap)
}

func NewClientMapByFile(name string)(ClientMap, error){
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	m := NewClientMap()
	//空文件
	if len(data) == 0 {
		return m, nil
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

func (cm ClientMap)Get(name string)(*Client, error){
	if s, ok := cm[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("cannot get client named %s, not found", name)
}

func (cm ClientMap)Save(name string)error{
	data, err := json.Marshal(&cm)
	if err != nil {
		return err
	}
	err = os.WriteFile(name, data, os.ModePerm)
	return err
}

func (cm ClientMap)ExistsByAddr(addr string)error{
	for _, client := range cm {
		if client.Address == addr {
			return nil
		}
	}
	return fmt.Errorf("cannot get client by addr %s, not found", addr)
}

func (cm ClientMap)Add(clt *Client){
	cm[clt.Name] = clt
}

func (cm ClientMap)AddClient(name string, conn net.Conn)error{
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}
	clt, err := cm.Get(name); 
	if err != nil {
		client := NewClient(name)
		client.Connection = conn
		client.Address = conn.RemoteAddr().String()
		cm.Add(&client)
		return nil
	}
	clt.Connection = conn
	clt.Address = conn.RemoteAddr().Network()
	return nil
}

func (cm ClientMap)GetProxy(client string, proxy string)(*Proxy, error){
	clt, err := cm.Get(client)
	if err != nil {
		return nil, err
	}
	prx, err := clt.ProxyMap.Get(proxy)
	return prx, err
}