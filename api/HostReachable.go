package api

import (
	"allproxy/config"
	"net"
	"os/exec"
	"strconv"
	"sync"
)

func SetHostReachable(proxyConfig *config.ProxyConfig, wg sync.WaitGroup) {
	go func() {
		cmd := exec.Command("ping", "-c", "1", proxyConfig.Hostname)
		if _, err := cmd.Output(); err == nil {
			conn, err := net.Dial("tcp", proxyConfig.Hostname+":"+strconv.Itoa(proxyConfig.Port))
			if err == nil {
				conn.Close()
				proxyConfig.HostReachable = true
			} else {
				proxyConfig.HostReachable = false
			}
		} else {
			proxyConfig.HostReachable = false
		}
		wg.Done()
	}()
}
