package config

import (
	"fmt"
	"net/http"
)

type ConfigProtocol string

const (
	Browser ConfigProtocol = "browser:"
	Grpc    ConfigProtocol = "grpc:"
	Http    ConfigProtocol = "http:"
	Https   ConfigProtocol = "https:"
	Log     ConfigProtocol = "log:"
	Mongo   ConfigProtocol = "mongo:"
	Redis   ConfigProtocol = "redis:"
	MySql   ConfigProtocol = "mysql:"
	Tcp     ConfigProtocol = "tcp:"
)

type ProxyConfig struct {
	IsSecure        bool           `json:"isSecure"`
	Path            string         `json:"path"`
	Protocol        ConfigProtocol `json:"protocol"`
	Hostname        string         `json:"hostname"`
	Port            int            `json:"port"`
	Recording       bool           `json:"recording"`
	HostReachable   bool           `json:"hostReachable"`
	LogProxyProcess string         `json:"logProxyProcess"`
	Server          *http.Server   `json:"_server"`
	Comment         string         `json:"comment"`
}

type ProxyConfigJson struct {
	Configs []*ProxyConfig `json:"configs"`
}

var Default = []*ProxyConfig{{
	IsSecure:      false,
	Protocol:      Browser,
	Path:          "/",
	Recording:     true,
	HostReachable: true,
}}

func (p *ProxyConfig) String() string {
	return fmt.Sprintf("%v\n", *p)
}
