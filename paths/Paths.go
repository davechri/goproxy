package paths

import (
	"log"
	"os"
	"path/filepath"
)

const (
	goproxyDataDir = "GOPROXY_DATA_DIR"
)

func dataDir() string {
	dir := os.Getenv(goproxyDataDir)
	if len(dir) == 0 {
		log.Panicln("GOPROXY_DATA_DIR environment variable must be set!")
	}
	return filepath.Clean(dir)
}

func ConfigJson() string {
	return filepath.Join(dataDir(), "config.json")
}

func ReplaceResponsesDir() string {
	return filepath.Join(dataDir(), "replace-responses")
}

func MakeCaDir() {
	if _, err := os.Stat(sslCaDir()); os.IsNotExist(err) {
		err := os.Mkdir(sslCaDir(), 0755)
		if err != nil {
			log.Panicln(err)
		}
		err = os.Mkdir(filepath.Join(sslCaDir(), "certs"), 0755)
		if err != nil {
			log.Panicln(err)
		}
		err = os.Mkdir(filepath.Join(sslCaDir(), "keys"), 0755)
		if err != nil {
			log.Panicln(err)
		}
	}
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
	oldName := filepath.Join(sslCaDir(), "certs/ca.pem")
	newName := filepath.Join(dataDir(), "ca.pem")
	os.Symlink(oldName, newName)
}

func ClientDir() string {
	return filepath.Join(dataDir(), "client")
}
