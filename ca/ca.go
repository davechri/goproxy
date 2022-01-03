package ca

import (
	"allproxy/paths"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	caPemName        = "ca.pem"
	caPrivateKeyName = "ca.private.key"
)

type ca struct {
	template *x509.Certificate
	key      *rsa.PrivateKey
}

var Ca ca

// Init CA
func InitCa() {
	log.Println("InitCa()")
	Ca.template = &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			CommonName:         "AllProxyCA",
			Organization:       []string{"AllProxy CA"},
			OrganizationalUnit: []string{"CA"},
			Country:            []string{"Internet"},
			Province:           []string{"Internet"},
			Locality:           []string{"Internet"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
		IsCA:      true,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageEmailProtection,
			x509.ExtKeyUsageTimeStamping,
		},
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageContentCommitment |
			x509.KeyUsageCertSign |
			x509.KeyUsageDataEncipherment |
			x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	if stat, err := os.Stat(filepath.Join(paths.SslCertsDir(), caPemName)); err == nil {
		// Read the private key from file (ca.private.key)
		f, err := os.Open(filepath.Join(paths.SslKeysDir(), caPrivateKeyName))
		if err != nil {
			log.Panicln(err)
		}
		buffer := make([]byte, stat.Size()*2)
		_, err = f.Read(buffer)
		if err != nil {
			log.Panic(err)
		}
		block, _ := pem.Decode(buffer)
		if block == nil {
			log.Panicf("pem.Decode failed for:\n %s", buffer)
		}
		Ca.key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		var err error
		Ca.key, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Panicln(err)
		}
		keyWriteToFile("ca", Ca.key)

		caBytes, err := x509.CreateCertificate(rand.Reader, Ca.template, Ca.template, &Ca.key.PublicKey, Ca.key)
		if err != nil {
			log.Panicln(err)
		}

		// CA Certificate in PEM format
		certWriteToFile("ca", caBytes)
	}
}

// Generate Server Certificate/Key
func NewServerCertKey(host string) (certFile string, keyFile string) {
	log.Printf("NewServerCertKey(%s)\n", host)
	if Ca.template == nil {
		log.Panicln("ca.NewCa() function was not called!")
	}

	certFile, keyFile = filepath.Join(paths.SslCertsDir(), host+".pem"), filepath.Join(paths.SslKeysDir(), host+".key")

	if _, err := os.Stat(certFile); err != nil {

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Panicln(err)
		}

		certTemplate := &x509.Certificate{
			SerialNumber: randomSerialNumber(),
			Subject: pkix.Name{
				Organization:       []string{"AllProxy Server Certificate"},
				OrganizationalUnit: []string{"AllProxy Server Certificate"},
				Country:            []string{"Internet"},
				Province:           []string{"Internet"},
				Locality:           []string{"Internet"},
				CommonName:         host,
			},
			DNSNames:              []string{host},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(10, 0, 0),
			IsCA:                  false,
			BasicConstraintsValid: true,
			SubjectKeyId:          subjectKeyId(privateKey),
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth,
				x509.ExtKeyUsageServerAuth,
			},
			KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, Ca.template, &privateKey.PublicKey, Ca.key)
		if err != nil {
			log.Panicln(err)
		}
		certWriteToFile(host, certBytes)

		keyWriteToFile(host, privateKey)
	}

	return certFile, keyFile
}

func certWriteToFile(name string, certBytes []byte) {
	fileName := filepath.Join(paths.SslCertsDir(), name+".pem")
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	os.WriteFile(fileName, certPEM.Bytes(), 0644)
}

func keyWriteToFile(name string, key *rsa.PrivateKey) {
	private := ""
	if name == "ca" {
		private = ".private" // only ca has ".private"
	}
	privateKeyFile := filepath.Join(paths.SslKeysDir(), name+private+".key")
	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	os.WriteFile(privateKeyFile, certPrivKeyPEM.Bytes(), 0644)

	publicKeyFile := filepath.Join(paths.SslKeysDir(), name+".public.key")
	certPublicKeyPEM := new(bytes.Buffer)
	pem.Encode(certPublicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	})
	os.WriteFile(publicKeyFile, certPublicKeyPEM.Bytes(), 0644)
}

func randomSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Panicln(err)
	}
	return serialNumber
}

func subjectKeyId(privKey *rsa.PrivateKey) []byte {
	pub := privKey.Public()
	// Subject Key Identifier support for end entity certificate.
	// https://tools.ietf.org/html/rfc3280#section-4.2.1.2
	pkixPub, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Panicln(err)
	}
	h := sha1.New()
	_, err = h.Write(pkixPub)
	if err != nil {
		log.Panicln(err)
	}
	keyID := h.Sum(nil)
	return keyID
}
