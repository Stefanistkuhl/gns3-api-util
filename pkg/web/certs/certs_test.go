package certs

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"sync"
	"testing"
	"time"
)

type mockCertStore struct {
	certPEM []byte
	keyPEM  []byte
	err     error
}

func (m *mockCertStore) GetCertificate(ctx context.Context) (certPEM, keyPEM []byte, err error) {
	return m.certPEM, m.keyPEM, m.err
}

func TestCertManagerSetGetCertificate(t *testing.T) {
	cm := &CertManager{}

	cert := cm.currentCert
	if cert != nil {
		t.Errorf("Initial certificate should be nil, got %v", cert)
	}

	testCert := &tls.Certificate{}

	cm.SetCertificate(testCert)

	hello := &tls.ClientHelloInfo{}
	got, err := cm.GetCertificate(hello)
	if err != nil {
		t.Errorf("GetCertificate() error = %v", err)
	}
	if got != testCert {
		t.Errorf("GetCertificate() = %v, want %v", got, testCert)
	}
}

func TestCertManagerLoadNewCert(t *testing.T) {
	cm := &CertManager{}

	certPEM, keyPEM := generateTestCertificate(t)

	err := cm.LoadNewCert(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadNewCert() error = %v", err)
	}

	hello := &tls.ClientHelloInfo{}
	cert, err := cm.GetCertificate(hello)
	if err != nil {
		t.Errorf("GetCertificate() after LoadNewCert error = %v", err)
	}
	if cert == nil {
		t.Error("GetCertificate() should return certificate after LoadNewCert")
	}
}

func generateTestCertificate(t *testing.T) (certPEM, keyPEM []byte) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return certPEM, keyPEM
}

func TestCertManagerLoadNewCertErrors(t *testing.T) {
	cm := &CertManager{}

	tests := []struct {
		name    string
		certPEM []byte
		keyPEM  []byte
		wantErr bool
	}{
		{"empty cert", []byte{}, []byte("key"), true},
		{"empty key", []byte("cert"), []byte{}, true},
		{"nil cert", nil, []byte("key"), true},
		{"nil key", []byte("cert"), nil, true},
		{"invalid cert", []byte("invalid cert"), []byte("invalid key"), true},
		{"mismatched pair", []byte("cert1"), []byte("key2"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.LoadNewCert(tt.certPEM, tt.keyPEM)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadNewCert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCertManagerConcurrentAccess(t *testing.T) {
	cm := &CertManager{}

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range numOperations {
				testCert := &tls.Certificate{}
				cm.SetCertificate(testCert)
			}
		}(i)
	}

	for range numGoroutines {
		wg.Go(func() {
			for range numOperations {
				hello := &tls.ClientHelloInfo{}
				_, err := cm.GetCertificate(hello)
				if err != nil {
					t.Errorf("GetCertificate() in goroutine error = %v", err)
				}
			}
		})
	}

	wg.Wait()
}

func TestCertManagerLoadFromEtcd(t *testing.T) {
	cm := &CertManager{}
	ctx := context.Background()

	t.Run("successful load", func(t *testing.T) {
		store := &mockCertStore{
			certPEM: []byte("cert"),
			keyPEM:  []byte("key"),
			err:     nil,
		}

		err := cm.LoadFromEtcd(ctx, store)
		if err == nil {
			t.Error("LoadFromEtcd() should fail with invalid cert/key")
		}
	})

	t.Run("store error", func(t *testing.T) {
		store := &mockCertStore{
			err: &mockError{"store error"},
		}

		err := cm.LoadFromEtcd(ctx, store)
		if err == nil {
			t.Error("LoadFromEtcd() should return store error")
		}
	})
}

func TestCertManagerWatchEtcd(t *testing.T) {
	cm := &CertManager{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &mockCertStore{
		certPEM: []byte("cert"),
		keyPEM:  []byte("key"),
		err:     nil,
	}

	notify := make(chan struct{}, 1)

	cm.WatchEtcd(ctx, store, notify)

	notify <- struct{}{}

	time.Sleep(10 * time.Millisecond)

	cancel()

	select {
	case notify <- struct{}{}:
		t.Log("Notification sent after context cancellation")
	default:
		t.Log("Notification channel full or closed")
	}
}

func TestCertStoreInterface(t *testing.T) {
	var _ CertStore = &mockCertStore{}
}

type mockError struct {
	msg string
}

func (m *mockError) Error() string {
	return m.msg
}

func TestCertManagerNilCertificate(t *testing.T) {
	cm := &CertManager{}

	hello := &tls.ClientHelloInfo{}
	cert, err := cm.GetCertificate(hello)
	if err != nil {
		t.Errorf("GetCertificate() with nil cert error = %v", err)
	}
	if cert != nil {
		t.Errorf("GetCertificate() with nil cert returned %v, want nil", cert)
	}
}

func TestCertManagerReplaceCertificate(t *testing.T) {
	cm := &CertManager{}

	cert1 := &tls.Certificate{}
	cm.SetCertificate(cert1)

	cert2 := &tls.Certificate{}
	cm.SetCertificate(cert2)

	hello := &tls.ClientHelloInfo{}
	got, err := cm.GetCertificate(hello)
	if err != nil {
		t.Errorf("GetCertificate() error = %v", err)
	}
	if got != cert2 {
		t.Errorf("GetCertificate() = %v, want %v", got, cert2)
	}
}
