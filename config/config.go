package config

type Server struct{
	Address string `json:"address,omitempty" yaml:"ip,omitempty"`
}

type Config struct {
	Server Server `json:"server,omitempty"`
}