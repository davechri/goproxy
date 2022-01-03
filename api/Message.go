package api

import "allproxy/config"

type MessageType int
type MessageProtocol string

const NoResponse = "No Response"
const (
	// eslint-disable-next-line no-unused-vars
	Request MessageType = 0
	// eslint-disable-next-line no-unused-vars
	Response MessageType = 1
	// eslint-disable-next-line no-unused-vars
	RequestAndResponse MessageType = 2
)

const (
	Http  MessageProtocol = "http:"
	Https MessageProtocol = "https:"
	Log   MessageProtocol = "log:"
	Mongo MessageProtocol = "mongo:"
	Redis MessageProtocol = "redis:"
	MySql MessageProtocol = "mysql:"
	Tcp   MessageProtocol = "tcp:"
)

type Message struct {
	Type            MessageType         `json:"type"`
	Timestamp       int                 `json:"timestamp"`
	SequenceNumber  int                 `json:"sequenceNumber"`
	RequestHeaders  map[string]string   `json:"requestHeaders"`
	ResponseHeaders map[string]string   `json:"responseHeaders"`
	Method          string              `json:"method"`
	Protocol        MessageProtocol     `json:"protocol"`
	Url             string              `json:"url"`
	Endpoint        string              `json:"endpoint"`
	RequestBody     interface{}         `json:"requestBody"`
	ResponseBody    interface{}         `json:"responseBody"`
	ClientIp        string              `json:"clientIp"`
	ServerHost      string              `json:"serverHost"`
	Path            string              `json:"path"`
	ElapsedTime     int                 `json:"elapsedTime"`
	Status          int                 `json:"status"`
	ProxyConfig     *config.ProxyConfig `json:"proxyConfig"`
}
