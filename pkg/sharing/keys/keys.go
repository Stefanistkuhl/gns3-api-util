package keys

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"os"
	"time"
)

type DeviceKey struct {
	Priv      ed25519.PrivateKey
	Pub       ed25519.PublicKey
	FP        string
	Path      string
	CreatedAt time.Time
}

type Options struct {
	Path string
}

func LoadOrCreate(opts Options) (*DeviceKey, error) {
	path := opts.Path
	if path == "" {
		var err error
		path, err = DefaultKeyPath()
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(path); err == nil {
		return Load(path)
	}
	return Create(path)
}

func Load(path string) (*DeviceKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	priv, createdAt, err := ParseEd25519PrivateKeyPEM(b)
	if err != nil {
		return nil, err
	}
	pub := priv.Public().(ed25519.PublicKey)
	fp := Fingerprint(pub)
	return &DeviceKey{
		Priv:      priv,
		Pub:       pub,
		FP:        fp,
		Path:      path,
		CreatedAt: createdAt,
	}, nil
}

func Create(path string) (*DeviceKey, error) {
	if err := os.MkdirAll(dir(path), 0o700); err != nil {
		return nil, err
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	pemBytes := MarshalEd25519PrivateKeyPEM(priv, time.Now().UTC())
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		return nil, err
	}
	pub := priv.Public().(ed25519.PublicKey)
	return &DeviceKey{
		Priv:      priv,
		Pub:       pub,
		FP:        Fingerprint(pub),
		Path:      path,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func Save(k *DeviceKey) error {
	if k == nil || k.Priv == nil {
		return errors.New("nil key")
	}
	if k.Path == "" {
		p, err := DefaultKeyPath()
		if err != nil {
			return err
		}
		k.Path = p
	}
	pemBytes := MarshalEd25519PrivateKeyPEM(k.Priv, k.CreatedAt)
	return os.WriteFile(k.Path, pemBytes, 0o600)
}

func Rotate(path string) (*DeviceKey, error) {
	return Create(path)
}

func dir(p string) string {
	i := len(p) - 1
	for i >= 0 && p[i] != '/' && p[i] != '\\' {
		i--
	}
	if i <= 0 {
		return "."
	}
	return p[:i]
}
