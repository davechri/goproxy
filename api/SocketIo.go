package api

import (
	"allproxy/config"
	"allproxy/paths"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	socketio "github.com/googollee/go-socket.io"
)

var cacheSocketId = "cache"

var windowSize = 500 // windows size - maximum outstanding messages
var maxOut = 2       // two message batches

func Start() *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		fmt.Println("SocketIo OnConnect() connected:", s.ID())
		s.SetContext("")
		proxyConfig := GetConfig()
		s.Emit("proxy config", proxyConfig) // send config to browser
		return nil
	})

	server.OnEvent("/", "proxy config", func(s socketio.Conn, proxyConfigs []*config.ProxyConfig) {
		fmt.Printf("SocketIo OnEvent \"proxy config\"\n%s\n", FmtConfig(proxyConfigs))
		saveConfig(proxyConfigs)

		// Make sure all matching connection based servers are closed.
		for i := range proxyConfigs {
			if proxyConfigs[i].Server != nil {
				closeAnyServerWithPort(proxyConfigs[i].Port)
			}
		}
		activateConfig(proxyConfigs, s)
	})

	server.OnEvent("/", "resend", func(
		s socketio.Conn,
		forwardProxy bool,
		method string,
		url string,
		message Message,
		body interface{},
	) {
		// resend(forwardProxy, method, url, message, body);
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("SocketIo OnError() meet error:", e)
	})

	server.OnDisconnect("/", func(socket socketio.Conn, reason string) {
		fmt.Println("SocketIo OnDisconnect()", reason)
		closeAnyServersWithSocket(socket.ID())
		socketIoMapDelete(socket.ID())
	})

	go server.Serve()
	http.Handle("/socket.io/", server)
	return server
}

func GetConfig() []*config.ProxyConfig {
	configJson := paths.ConfigJson()
	if data, err := os.ReadFile(configJson); err == nil {
		var proxyConfigJson config.ProxyConfigJson
		if err := json.Unmarshal(data, &proxyConfigJson); err != nil {
			log.Panicln(err)
		}
		return proxyConfigJson.Configs
	} else {
		return config.Default
	}
}

func FmtConfig(proxyConfigJson []*config.ProxyConfig) string {
	str, err := json.MarshalIndent(proxyConfigJson, "", "  ")
	if err != nil {
		log.Panicln(err)
	}
	return string(str)
}

func UpdateHostReachable() {
	var wg sync.WaitGroup
	proxyConfigs := GetConfig()
	for i := range proxyConfigs {
		if proxyConfigs[i].Protocol == config.Browser || proxyConfigs[i].Protocol == config.Log {
			proxyConfigs[i].HostReachable = true
		} else {
			wg.Add(1)
			SetHostReachable(proxyConfigs[i], wg)
		}
	}
	wg.Wait()
}

func saveConfig(proxyConfigs []*config.ProxyConfig) {
	// Cache the config, to configure the proxy on the next start up prior
	// to receiving the config from the browser.
	proxyConfigJson := config.ProxyConfigJson{Configs: proxyConfigs}
	if data, err := json.MarshalIndent(proxyConfigJson, "", "  "); err == nil {
		os.WriteFile(paths.ConfigJson(), data, 0644)
	} else {
		log.Panicln(err)
	}
}

func activateConfig(proxyConfigs []*config.ProxyConfig, socket socketio.Conn) {
	// for _, proxyConfig := range proxyConfigs {
	// if proxyConfig.protocol == config.Log {
	// 	new LogProxy(proxyConfig);
	// } else if proxyConfig.protocol == config.Grpc && USE_HTTP2) {
	// 	GrpcProxy.reverseProxy(proxyConfig);
	// } else if (
	// 	proxyConfig.protocol !== 'http:' &&
	// 	proxyConfig.protocol !== 'https:' &&
	// 	proxyConfig.protocol !== 'browser:'
	// ) {
	// 	// eslint-disable-next-line no-new
	// 	new TcpProxy(proxyConfig);
	// }
	// }

	key := cacheSocketId
	if socket != nil {
		key = socket.ID()
	}
	socketIoMapAdd(
		key,
		&socketIoInfo{
			socket:          socket,
			configs:         proxyConfigs,
			remainingWindow: windowSize,
		},
	)
	if socket != nil {
		closeAnyServersWithSocket(cacheSocketId)
		socketIoMapDelete(cacheSocketId)
	}
}

// Close 'any:' protocol servers that are running for the browser owning the socket
func closeAnyServersWithSocket(socketId string) {
	// for key, socketInfo := range socketIoMap {
	// 	if socketId != key {
	// 		continue
	// 	}
	// for _, proxyConfigs := range socketInfo.configs {
	// if proxyConfig.protocol == config.Log {
	// 	LogProxy.destructor(proxyConfig);
	// } else if proxyConfig.protocol == config.Grpc {
	// 	GrpcProxy.destructor(proxyConfig);
	// } else if proxyConfig._server {
	// 	TcpProxy.destructor(proxyConfig);
	// }
	// }
	// }
}

// Close 'any:' protocol servers the specified listening port
func closeAnyServerWithPort(port int) {
	// for _, socketInfo := range socketIoMap {
	// for _, proxyConfig := range socketInfo.configs {
	// if proxyConfig._server && proxyConfig.port == port {
	// 	if proxyConfig.protocol == config.Grpc && USE_HTTP2 {
	// 		GrpcProxy.destructor(proxyConfig)
	// 	} else {
	// 		TcpProxy.destructor(proxyConfig)
	// 	}
	// }
	// }
	// }
}
