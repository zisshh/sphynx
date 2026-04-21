package models

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"
)

type SSLCertificate struct {
	CertPath string
	KeyPath  string
	Port     int
}

var Certificates = map[int]*SSLCertificate{} // Maps ports to SSL certificates

func UploadCertificate(certPath, keyPath string, port int) error {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return errors.New("certificate file does not exist")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return errors.New("key file does not exist")
	}
	Certificates[port] = &SSLCertificate{CertPath: certPath, KeyPath: keyPath, Port: port}
	return nil
}

func GetCertificate(port int) (*SSLCertificate, error) {
	if cert, exists := Certificates[port]; exists {
		return cert, nil
	}
	return nil, fmt.Errorf("no certificate found for port %d", port)
}

// GenerateCertificate creates a self-signed certificate and saves it as a file
func GenerateCertificate(port int, commonName string, days int) (string, string, error) {
	certDir := fmt.Sprintf("certs/%d", port)
	certPath := fmt.Sprintf("%s/cert.crt", certDir)
	keyPath := fmt.Sprintf("%s/key.key", certDir)

	// Ensure the directory exists
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create directory for port %d: %v", port, err)
	}

	// Generate private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %v", err)
	}

	// Define certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, days),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{commonName},
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %v", err)
	}

	// Save the certificate and private key
	if err := writePEM(certPath, "CERTIFICATE", certDER); err != nil {
		return "", "", err
	}
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %v", err)
	}
	if err := writePEM(keyPath, "EC PRIVATE KEY", keyBytes); err != nil {
		return "", "", err
	}

	return certPath, keyPath, nil
}


// Helper function to write PEM files
func writePEM(filePath, pemType string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{Type: pemType, Bytes: data})
	if err != nil {
		return fmt.Errorf("failed to write PEM block to file: %v", err)
	}

	return nil
}

// RenewCertificate renews a cert
func RenewCertificate(port int, commonName string, days int) (string, string, error) {
	certDir := fmt.Sprintf("certs/%d", port)
	certPath := fmt.Sprintf("%s/cert.crt", certDir)
	keyPath := fmt.Sprintf("%s/key.key", certDir)

	// Verify existing certificate
	if err := verifyCertificateChain(certPath); err != nil {
		return "", "", fmt.Errorf("existing certificate validation failed: %v", err)
	}

	// Backup existing certificate and key
	backupCert := fmt.Sprintf("%s/cert_backup.crt", certDir)
	backupKey := fmt.Sprintf("%s/key_backup.key", certDir)
	if err := backupFile(certPath, backupCert); err != nil {
		return "", "", fmt.Errorf("failed to backup certificate: %v", err)
	}
	if err := backupFile(keyPath, backupKey); err != nil {
		return "", "", fmt.Errorf("failed to backup key: %v", err)
	}

	// Generate a new certificate
	return GenerateCertificate(port, commonName, days)
}

// veriffy certificate chain
func verifyCertificateChain(certPath string) error {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Validate the certificate chain
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(data) {
		return fmt.Errorf("failed to add certificate to root pool")
	}
	opts := x509.VerifyOptions{Roots: roots}
	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate chain validation failed: %v", err)
	}

	return nil
}


// Helper function to back up a file
func backupFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %v", err)
	}

	if err := os.WriteFile(dest, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %v", err)
	}

	return nil
}

// IsCertificateExpiring checks if the certificate is expiring within the threshold days
func IsCertificateExpiring(certPath string, thresholdDays int) (bool, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return false, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Check expiration
	return time.Now().AddDate(0, 0, thresholdDays).After(cert.NotAfter), nil
}

// Certificate rotation
func RotateCertificates() {
	for port, cert := range Certificates {
		// Verify the certificate chain
		if err := verifyCertificateChain(cert.CertPath); err != nil {
			fmt.Printf("Certificate for port %d is invalid: %v\n", port, err)
			continue
		}

		// Check if the certificate is expiring within the next 30 days
		expiring, err := IsCertificateExpiring(cert.CertPath, 30)
		if err != nil {
			fmt.Printf("Error checking certificate for port %d: %v\n", port, err)
			continue
		}

		if expiring {
			fmt.Printf("Certificate for port %d is expiring, renewing...\n", port)
			_, _, err := RenewCertificate(port, "example.com", 365)
			if err != nil {
				fmt.Printf("Failed to renew certificate for port %d: %v\n", port, err)
				continue
			}

			// Update the Certificates map
			certDir := fmt.Sprintf("certs/%d", port)
			Certificates[port] = &SSLCertificate{
				CertPath: fmt.Sprintf("%s/cert.crt", certDir),
				KeyPath:  fmt.Sprintf("%s/key.key", certDir),
				Port:     port,
			}
			fmt.Printf("Certificate for port %d successfully renewed.\n", port)
		}
	}
}


// Scheduler for rotation
func StartCertificateRotation() {
	go func() {
		for {
			RotateCertificates()
			time.Sleep(24 * time.Hour) // Daily rotation check
		}
	}()
}
