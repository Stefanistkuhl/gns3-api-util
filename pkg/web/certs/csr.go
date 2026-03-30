package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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

func GenerateNodeKeyAndCSR(tlsDir, nodeName string, domains []string) (csrPEM []byte, err error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, marshalErr := x509.MarshalPKCS8PrivateKey(privKey)
	if marshalErr != nil {
		return nil, marshalErr
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	if mkdirErr := os.MkdirAll(tlsDir, 0o700); mkdirErr != nil {
		return nil, mkdirErr
	}
	if writeErr := os.WriteFile(filepath.Join(tlsDir, "node.key"), keyPEM, 0o600); writeErr != nil {
		return nil, writeErr
	}

	subj := pkix.Name{
		CommonName:   nodeName,
		Organization: []string{"gns3util-cluster"},
	}

	// TODO(ingress): accept a stable hostname/domain from config (e.g. a load
	// balancer VIP, k8s service DNS, or a *.local mDNS name) and include it in
	// DNSNames so that clients can always connect via that name regardless of
	// which interface IP the server is currently reachable on.  Right now
	// IPAddresses is a snapshot of all local IPs at cert-generation time, which
	// means any IP that wasn't up at that moment (e.g. a second NIC, a new DHCP
	// lease, a WiFi interface) causes hostname-verification failures on clients
	// and forces the InsecureSkipVerify workaround in pkg/api/client.go.
	if len(domains) == 0 {
		// TODO(ingress): replace this "localhost" fallback with the cluster's
		// configured ingress hostname so nodes always get a meaningful SAN even
		// without explicit domain config.
		domains = []string{"localhost"}
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		DNSNames:           domains,
		IPAddresses:        nwutils.GetLocalIPs(),
	}

	csrBytes, createErr := x509.CreateCertificateRequest(rand.Reader, &template, privKey)
	if createErr != nil {
		return nil, createErr
	}

	csrPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csrPEM, nil
}

func SignCSR(csrPEM []byte, caCert *x509.Certificate, caPrivKey any) ([]byte, error) {
	block, _ := pem.Decode(csrPEM)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("failed to decode PEM block containing CSR")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}
	if signatureCheckErr := csr.CheckSignature(); signatureCheckErr != nil {
		return nil, fmt.Errorf("invalid CSR signature: %w", signatureCheckErr)
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
		DNSNames:              csr.DNSNames,
		IPAddresses:           csr.IPAddresses,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, csr.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes}), nil
}
