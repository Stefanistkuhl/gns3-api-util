package transport

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/quic-go/quic-go"
)

type Server struct {
	TLS       *tls.Config
	ServerKey ed25519.PrivateKey // used to derive SAS (server public key)
}

func (s *Server) Listen(
	ctx context.Context,
	addr string,
) (net.Addr, <-chan error, func() error, error) {
	if s.TLS == nil {
		return nil, nil, nil, fmt.Errorf("TLS config required")
	}
	ln, err := quic.ListenAddr(addr, s.TLS, &quic.Config{})
	if err != nil {
		return nil, nil, nil, err
	}

	errs := make(chan error, 1)
	go func() {
		defer close(errs)
		for {
			c, err := ln.Accept(ctx)
			if err != nil {
				// Accept returns error when listener is closed or ctx canceled.
				errs <- err
				return
			}
			go handleConn(ctx, c, s.ServerKey)
		}
	}()

	closer := func() error { return ln.Close() }
	return ln.Addr(), errs, closer, nil
}

func handleConn(ctx context.Context, c *quic.Conn, serverPriv ed25519.PrivateKey) {
	// 1) Accept the bidirectional control stream
	ctrl, err := c.AcceptStream(ctx)
	if err != nil {
		_ = c.CloseWithError(0, "accept stream failed")
		return
	}

	// 2) Hello exchange
	var cli Hello
	if err := ReadJSON(ctx, ctrl, &cli); err != nil {
		_ = c.CloseWithError(0, "bad hello")
		return
	}
	if err := WriteJSON(ctx, ctrl, Hello{Label: "server", FP: "unknown"}); err != nil {
		_ = c.CloseWithError(0, "write hello failed")
		return
	}

	// 3) SAS nonce exchange
	var clientSAS SASMsg
	if err := ReadJSON(ctx, ctrl, &clientSAS); err != nil {
		_ = c.CloseWithError(0, "read client nonce")
		return
	}
	serverNonce, err := NewNonce()
	if err != nil {
		_ = c.CloseWithError(0, "nonce gen failed")
		return
	}
	if err := WriteJSON(ctx, ctrl, SASMsg{Nonce: serverNonce}); err != nil {
		_ = c.CloseWithError(0, "write server nonce")
		return
	}

	serverPub := serverPriv.Public().(ed25519.PublicKey)
	if words, err := DerivePGPWordsSimple(serverPub, clientSAS.Nonce, serverNonce, 3); err == nil {
		fmt.Printf("Verify code: %s\n", FormatSAS(words))
	}

	// 4) Receive Offer (list of files)
	var offer OfferMsg
	if err := ReadJSON(ctx, ctrl, &offer); err != nil {
		_ = c.CloseWithError(0, "bad offer")
		return
	}
	fmt.Printf("Incoming files (%d):\n", len(offer.Files))
	for _, fm := range offer.Files {
		fmt.Printf("  - %s (%d bytes)\n", fm.Rel, fm.Size)
	}

	// 5) Destination directory: ~/.gns3
	home, _ := os.UserHomeDir()
	dst := filepath.Join(home, ".gns3")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		_ = c.CloseWithError(0, "mkdir failed")
		return
	}

	// 6) Receive the advertised number of files via unidirectional streams
	if err := ReceiveFiles(ctx, c, dst, len(offer.Files)); err != nil {
		_ = c.CloseWithError(0, "recv failed")
		return
	}
	fmt.Println("Receive complete.")

	// 7) Mirror client shutdown: when client closes, close from server side too.
	<-c.Context().Done()
	// CloseWithError is idempotent; ignore returned error.
	_ = c.CloseWithError(0, "ok")
}
