package keys

import (
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const pemType = "ED25519 PRIVATE KEY"

func MarshalEd25519PrivateKeyPEM(priv ed25519.PrivateKey, createdAt time.Time) []byte {
	headers := map[string]string{
		"Created": strconv.FormatInt(createdAt.Unix(), 10),
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:    pemType,
		Bytes:   []byte(priv),
		Headers: headers,
	})
}

func ParseEd25519PrivateKeyPEM(b []byte) (ed25519.PrivateKey, time.Time, error) {
	blk, _ := pem.Decode(b)
	if blk == nil || blk.Type != pemType {
		return nil, time.Time{}, errors.New("invalid or missing ed25519 PEM")
	}
	priv := ed25519.PrivateKey(blk.Bytes)
	if len(priv) != ed25519.PrivateKeySize {
		return nil, time.Time{}, fmt.Errorf("unexpected ed25519 key size: %d", len(priv))
	}
	var created time.Time
	if v, ok := blk.Headers["Created"]; ok {
		if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
			created = time.Unix(sec, 0).UTC()
		}
	}
	if created.IsZero() {
		created = time.Now().UTC()
	}
	return priv, created, nil
}

