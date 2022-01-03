package http

import (
	"encoding/json"
	"goproxy/api"
	"goproxy/config"
	"goproxy/dns"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HttpMessage struct {
	EmitCount       int
	StartTime       int64
	MessageProtocol api.MessageProtocol
	ProxyConfig     *config.ProxyConfig
	PipelineSeqNum  int32
	SequenceNumber  int
	RemoteAddress   string
	Method          string
	Url             string
	ReqHeaders      http.Header
	ReqBody         interface{}
}

func (hm *HttpMessage) Deadline() (deadline time.Time, ok bool) {
	return
}
func (hm *HttpMessage) Done() <-chan struct{}             { return nil }
func (hm *HttpMessage) Err() error                        { return nil }
func (hm *HttpMessage) Value(key interface{}) interface{} { return nil }

func NewHttpMessage(
	messageProtocol api.MessageProtocol,
	proxyConfig *config.ProxyConfig,
	pipelineSeqNum int32,
	sequenceNumber int,
	remoteAddress string,
	method string,
	url string,
	reqHeaders http.Header,
	reqBody interface{},
) *HttpMessage {
	hm := HttpMessage{
		StartTime:       time.Now().Unix(),
		MessageProtocol: messageProtocol,
		ProxyConfig:     proxyConfig,
		PipelineSeqNum:  pipelineSeqNum,
		SequenceNumber:  sequenceNumber,
		RemoteAddress:   remoteAddress,
		Method:          method,
		Url:             url,
		ReqHeaders:      reqHeaders,
		ReqBody:         reqBody,
	}
	return &hm
}

func (hm *HttpMessage) EmitMessageToBrowser(
	resStatus int,
	resHeaders http.Header,
	resBody interface{},
) {
	reqBodyJson := parseBody(hm.ReqBody)
	var resBodyJson interface{}
	if resBody == api.NoResponse {
		resBodyJson = resBody
	} else {
		resBodyJson = parseBody(resBody)
	}
	host := "Unknown"
	if hm.ProxyConfig != nil {
		host = getHostPort(hm.ProxyConfig, hm.ReqHeaders)
	}

	messageType := api.Request
	if resBody != api.NoResponse {
		if hm.EmitCount == 0 {
			messageType = api.RequestAndResponse
		} else {
			messageType = api.Response
		}
	}
	message := api.Message{
		Type:            messageType,
		Timestamp:       int(time.Now().Unix() - hm.StartTime),
		SequenceNumber:  hm.SequenceNumber,
		RequestHeaders:  removeDupHeaders(hm.ReqHeaders),
		ResponseHeaders: removeDupHeaders(resHeaders),
		Method:          hm.Method,
		Protocol:        hm.MessageProtocol,
		Url:             hm.Url,
		Endpoint:        getHttpEndpoint(hm.Method, hm.Url, reqBodyJson),
		RequestBody:     reqBodyJson,
		ResponseBody:    resBodyJson,
		ClientIp:        hm.RemoteAddress,
		ServerHost:      dns.ResolveIp(host),
		Path:            hm.ProxyConfig.Path,
		ElapsedTime:     int(hm.StartTime - time.Now().Unix()),
		Status:          resStatus,
		ProxyConfig:     hm.ProxyConfig,
	}

	api.EmitMessageToBrowser(messageType, &message, hm.ProxyConfig)
	hm.EmitCount++
}

func removeDupHeaders(header http.Header) map[string]string {
	out := make(map[string]string)
	for key, values := range header {
		out[key] = values[0]
	}
	return out
}

func parseBody(body interface{}) interface{} {
	switch v := body.(type) {
	case string:
		var j map[string]interface{}
		err := json.Unmarshal([]byte(v), &j)
		if err != nil {
			return v
		}
		return j
	default:
		return body
	}
}

func getHostPort(proxyConfig *config.ProxyConfig, reqHeaders http.Header) string {
	if len(proxyConfig.Hostname) > 0 {
		host := proxyConfig.Hostname
		if proxyConfig.Port != 0 {
			host += ":" + strconv.Itoa(proxyConfig.Port)
		}
		return host
	} else {
		host := ""
		if host = reqHeaders.Get("host"); len(host) == 0 {
			return "Unknown"
		}
		return host
	}
}

func getHttpEndpoint(method string, url string, requestBody interface{}) string {
	endpoint := strings.Split(url, "?")[0]
	tokens := strings.Split(endpoint, "/")
	if len(tokens) > 0 {
		endpoint = tokens[len(tokens)-1]
	}

	// This is an id?
	if _, err := strconv.Atoi(endpoint); err == nil {
		endpoint = tokens[len(tokens)-2] + "/" + tokens[len(tokens)-1]
	}

	appendOperation := func(v interface{}) {
		switch m := v.(type) {
		case map[string]string:
			if operation, ok := m["operationName"]; ok {
				if len(endpoint) > 0 {
					endpoint += ","
					endpoint += " " + operation
				}
			}
		}
	}

	// GraphQL?
	if method != "OPTIONS" &&
		(strings.HasSuffix(url, "/graphql") || strings.HasSuffix(url, "/graphql-public")) {
		endpoint = ""
		switch v := requestBody.(type) {
		case map[string]interface{}:
			appendOperation(v)
		case []interface{}:
			for i := range v {
				appendOperation(v[i])
			}
		}

		tag := "GQL"
		if strings.HasSuffix(url, "/graphql-public") {
			tag = "GQLP"
		}
		endpoint = " " + tag + endpoint
	}
	return endpoint
}
