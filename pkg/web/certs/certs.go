package certs

import (
	"context"
	"crypto/tls"
	"log"
	"sync"
)

type CertManager struct {
	mu          sync.RWMutex
	currentCert *tls.Certificate
}

func (cm *CertManager) SetCertificate(cert *tls.Certificate) {
	cm.mu.Lock()
	cm.currentCert = cert
	cm.mu.Unlock()
}

func (cm *CertManager) LoadNewCert(certBytes, keyBytes []byte) error {
	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return err
	}

	cm.mu.Lock()
	cm.currentCert = &cert
	cm.mu.Unlock()
	return nil
}

func (cm *CertManager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.currentCert, nil
}

func (cm *CertManager) LoadFromEtcd(ctx context.Context, store CertStore) error {
	certPEM, keyPEM, err := store.GetCertificate(ctx)
	if err != nil {
		return err
	}
	return cm.LoadNewCert(certPEM, keyPEM)
}

func (cm *CertManager) WatchEtcd(ctx context.Context, store CertStore, notify <-chan struct{}) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-notify:
				log.Println("Certificate update detected, reloading...")
				if err := cm.LoadFromEtcd(ctx, store); err != nil {
					log.Printf("Failed to reload cert: %v", err)
				} else {
					log.Println("Certificate reloaded successfully")
				}
			}
		}
	}()
}

type CertStore interface {
	GetCertificate(ctx context.Context) (certPEM, keyPEM []byte, err error)
}
