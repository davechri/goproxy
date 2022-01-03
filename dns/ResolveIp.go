package dns

import (
	"log"
	"net"
	"sort"
	"strings"
)

func ResolveIp(ipPort string) string {
	log.Println("ResolveIp()", ipPort)
	hostPort := strings.Split(ipPort, ":")
	domains, err := net.LookupAddr(hostPort[0])
	if err != nil {
		return ipPort
	}
	sort.Strings(domains)
	host := strings.Split(domains[0], ".")[0]
	return host
}
