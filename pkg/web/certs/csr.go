package certs

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/0xveya/gns3util/pkg/utils/nwutils"
)

func GenerateNodeKeyAndCSR(tlsDir, nodeName string) (csrPEM []byte, err error) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	if err := os.MkdirAll(tlsDir, 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(tlsDir, "node.key"), keyPEM, 0600); err != nil {
		return nil, err
	}

	subj := pkix.Name{
		CommonName:   nodeName,
		Organization: []string{"gns3util-cluster"},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.PureEd25519,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privKey)
	if err != nil {
		return nil, err
	}

	csrPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csrPEM, nil
}

func SignCSR(csrPEM []byte, caCert *x509.Certificate, caPrivKey ed25519.PrivateKey) ([]byte, error) {
	block, _ := pem.Decode(csrPEM)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("failed to decode PEM block containing CSR")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}
	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("invalid CSR signature: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{"localhost"},
		IPAddresses:           nwutils.GetLocalIPs(),
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, csr.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes}), nil
}
