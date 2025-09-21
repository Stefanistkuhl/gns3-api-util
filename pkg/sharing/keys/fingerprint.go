package keys

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base32"
	"strings"
)

func Fingerprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return enc.EncodeToString(sum[:])
}

func ShortFingerprint(fp string) string {
	var b strings.Builder
	for i := 0; i < len(fp); i += 4 {
		end := i + 4
		end = min(end, len(fp))
		if i > 0 {
			b.WriteByte('-')
		}
		b.WriteString(fp[i:end])
	}
	return b.String()
}
