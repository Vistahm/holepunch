package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// GenerateCert creates a self-signed TLS certificate and key.
// It includes the provided IP addresses so the browser can match them.
func GenerateCert(ips []net.IP, certFile, keyFile string) error {
	// Generate ECDSA private key
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	// Create certificate template
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"HolePunch Self-Signed"},
			CommonName:   "holepunch.local",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add IP addresses to the certificate
	for _, ip := range ips {
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	// Write private key
	keyFileHandle, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	defer keyFileHandle.Close()

	keyBytes, _ := x509.MarshalECPrivateKey(key)
	pem.Encode(keyFileHandle, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	// Write certificate
	certFileHandle, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("create cert file: %w", err)
	}
	defer certFileHandle.Close()

	pem.Encode(certFileHandle, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return nil
}

// LoadConfig loads or generates a TLS config.
func LoadConfig(certFile, keyFile string) (*tls.Config, error) {
	// Generate if files don't exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		ips, _ := getLocalIPs()
		if err := GenerateCert(ips, certFile, keyFile); err != nil {
			return nil, err
		}
		fmt.Printf("🔐 Generated self-signed certificate (valid 1 year)\n")
	}

	// Load the certificate pair
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func getLocalIPs() ([]net.IP, error) {
	var ips []net.IP

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					ips = append(ips, ip4)
				}
			}
		}
	}

	return ips, nil
}
