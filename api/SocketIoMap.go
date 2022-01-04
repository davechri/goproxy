package api

import (
	"encoding/json"
	"goproxy/config"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"

	socketio "github.com/googollee/go-socket.io"
)

type socketIoInfo struct {
	socket          socketio.Conn
	configs         []*config.ProxyConfig
	seqNum          int
	remainingWindow int
	messagesOut     int
	queuedMessages  []*Message
}

var socketIoMap = sync.Map{}

func socketIoMapAdd(key string, socketInfo *socketIoInfo) {
	socketIoMap.Store(key, socketInfo)
}

func socketIoMapDelete(key string) {
	socketIoMap.Delete(key)
}

func isMatch(needle string, haystack string) bool {
	if strings.Contains(haystack, ".*") {
		b, err := regexp.MatchString(needle, haystack)
		if err != nil {
			log.Fatalln(err)
		}
		return b
	} else {
		return strings.HasPrefix(haystack, needle)
	}
}

/**
 * Find proxy config matching URL
 * @params isSecure
 * @params clientHostName
 * @param {*} reqUrl
 * @param isForwardProxy
 * @returns ProxyConfig
 */
func FindProxyConfigMatchingURL(
	isSecure bool,
	clientHostName string,
	reqUrl *url.URL,
	isForwardProxy bool,
) *config.ProxyConfig {
	reqUrlPath := strings.ReplaceAll(reqUrl.Path, "//", "/")
	var matchingProxyConfig *config.ProxyConfig

	// Find matching proxy configuration
	socketIoMap.Range(func(_ interface{}, value interface{}) bool {
		for _, proxyConfig := range value.(*socketIoInfo).configs {
			if proxyConfig.Protocol != config.Http && proxyConfig.Protocol != config.Browser {
				continue
			}
			if (isMatch(proxyConfig.Path, reqUrlPath) ||
				isMatch(proxyConfig.Path, clientHostName+reqUrlPath)) &&
				isForwardProxy == (proxyConfig.Protocol == config.Browser) {
				if matchingProxyConfig == nil || len(proxyConfig.Path) > len(matchingProxyConfig.Path) {
					matchingProxyConfig = proxyConfig
				}
			}
		}
		return true
	})

	return matchingProxyConfig
}

/**
 * Emit message to browser.
 * @param {*} message
 * @param {*} proxyConfig
 */
func EmitMessageToBrowser(messageType MessageType, message *Message, inProxyConfig *config.ProxyConfig) {
	// log.Println("SocketIo EmitMessageToBrowser()", socketIoMap)
	message.Type = messageType
	path := ""
	if inProxyConfig != nil {
		path = inProxyConfig.Path
	}
	var currentSocketId string
	emitted := false
	socketIoMap.Range(func(key interface{}, value interface{}) bool {
		for _, proxyConfig := range value.(*socketIoInfo).configs {
			if inProxyConfig == nil ||
				(proxyConfig.Path == path && inProxyConfig.Protocol == proxyConfig.Protocol) {
				// Don't emit to same socket again
				if key == currentSocketId {
					continue
				}
				// Recording is turned off?
				if !proxyConfig.Recording {
					continue
				}
				currentSocketId = key.(string)
				message.ProxyConfig = proxyConfig
				if value.(*socketIoInfo).socket != nil {
					emitMessageWithFlowControl([]*Message{message}, value.(*socketIoInfo), currentSocketId)
					emitted = true
				}
				if inProxyConfig == nil {
					break
				}
			}
		}
		return true
	})
	// if !emitted {
	// 	log.Println(message.SequenceNumber, "no browser socket to emit to", message.Url)
	// }
}

func emitMessageWithFlowControl(messages []*Message, socketInfo *socketIoInfo, socketId string) {
	// log.Println("SocketIo emitMessageWithFlowControl", socketId)
	if socketInfo.remainingWindow == 0 || socketInfo.messagesOut >= maxOut {
		socketInfo.queuedMessages = append(socketInfo.queuedMessages, messages...)
	} else {
		if socketInfo.socket != nil {
			batchCount := len(messages)
			socketInfo.remainingWindow -= batchCount
			socketInfo.seqNum++
			socketInfo.messagesOut++
			messagesBytes, err := json.MarshalIndent(messages, "", "  ")
			if err != nil {
				log.Panicln(err)
			}
			// log.Println("SocketIo emitMessageWithFlowControl() messages:")
			socketInfo.socket.Emit(
				"reqResJson",
				string(messagesBytes),
				len(socketInfo.queuedMessages),
				// callback:
				func(response string) {
					socketInfo.messagesOut--
					socketInfo.remainingWindow += batchCount

					log.Println(
						"out=", socketInfo.messagesOut,
						"win=", socketInfo.remainingWindow,
						"queued=", len(socketInfo.queuedMessages),
						response)

					count := len(socketInfo.queuedMessages)
					if socketInfo.remainingWindow < len(socketInfo.queuedMessages) {
						count = socketInfo.remainingWindow
					}
					if count > 0 {
						emitMessageWithFlowControl(
							socketInfo.queuedMessages[0:count],
							socketInfo,
							socketId,
						)
						socketInfo.queuedMessages = socketInfo.queuedMessages[count:]
					}
				},
			)
		}
	}
}
