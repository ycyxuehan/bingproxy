package proxy

import (
	"encoding/json"
	"time"
)

type CommandType string

const (
	Connect CommandType = "Connect"
	Release CommandType = "Release"
	Register CommandType = "Register"
	KeepAlive CommandType = "KeepAlive"
)

type Command struct {
	Type       CommandType   `json:"type,omitempty"`
	Name       string        `json:"hash,omitempty"`
	RemoteAddr string        `json:"remoteAddr,omitempty"`
	LocalAddr  string        `json:"localAddr,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty"`
}

func NewCommand(name string, t CommandType, remote string, local string)Command{
	d := Command{
		Type: t,
		Name: name,
		RemoteAddr: remote,
		LocalAddr: local,
	}
	return d
}

func NewBytesCommand(data []byte) (*Command, error) {
	d := Command{}
	err := json.Unmarshal(data, &d)
	return &d, err
}

func (cmd *Command)Bytes()([]byte, error){
	return json.Marshal(cmd)
}

func (cmd *Command)String()string{
	b, _ := cmd.Bytes()
	return string(b)
}