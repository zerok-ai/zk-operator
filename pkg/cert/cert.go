package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
)

func InitializeKeysAndCertificates(cfg config.WebhookConfig) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {

	dnsNames := []string{
		cfg.Service,
		cfg.Service + "." + cfg.Namespace,
		cfg.Service + "." + cfg.Namespace + ".svc",
	}
	commonName := cfg.Service + "." + cfg.Namespace + ".svc"

	// get CA PEM
	org := "zerok"
	caPEM, serverCertPEM, serverCertKeyPEM, err := generateCert([]string{org}, dnsNames, commonName)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate certificate: %v.\n", err)
		return nil, nil, nil, fmt.Errorf(msg)
	}

	return caPEM, serverCertPEM, serverCertKeyPEM, nil
}

func generateCert(orgs, dnsNames []string, commonName string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(int64(time.Now().Day())),
		Subject:               pkix.Name{Organization: orgs},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	serverCertPEM, serverPrivateKeyPEM, err := getServerCertPEM(orgs, dnsNames, commonName, ca, caPrivateKey)

	if err != nil {
		return nil, nil, nil, err
	}

	return caPEM, serverCertPEM, serverPrivateKeyPEM, nil
}

func getServerCertPEM(orgs, dnsNames []string, commonName string, parentCa *x509.Certificate, parentPrivateKey *rsa.PrivateKey) (*bytes.Buffer, *bytes.Buffer, error) {
	serverCert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1024),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: orgs,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCert, parentCa, &serverPrivateKey.PublicKey, parentPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	serverCertPEM := new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivateKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(serverPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})

	return serverCertPEM, serverPrivateKeyPEM, nil
}
