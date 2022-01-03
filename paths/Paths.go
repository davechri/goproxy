package paths

import (
	"log"
	"os"
	"path/filepath"
)

const (
	allProxyDataDir = "ALLPROXY_DATA_DIR"
)

func dataDir() string {
	dir := os.Getenv(allProxyDataDir)
	if len(dir) == 0 {
		log.Panicln("ALLPROXY_DATA_DIR environment variable must be set!")
	}
	return filepath.Clean(dir)
}

func ConfigJson() string {
	return filepath.Join(dataDir(), "config.json")
}

func ReplaceResponsesDir() string {
	return filepath.Join(dataDir(), "replace-responses")
}

func sslCaDir() string {
	dir := filepath.Join(dataDir(), ".http-mitm-proxy")
	return dir
}

func SslCertsDir() string {
	return filepath.Join(sslCaDir(), "certs")
}

func SslKeysDir() string {
	return filepath.Join(sslCaDir(), "keys")
}

func MakeCaPemSymLink() {
	oldName := filepath.Join(dataDir(), ".http-mitm-proxy/certs/ca.pem")
	newName := filepath.Join(dataDir(), "ca.pem")
	os.Symlink(oldName, newName)
}

func ClientDir() string {
	return filepath.Join(dataDir(), "client")
}
