package transport

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

type Server struct {
	TLS       *tls.Config
	ServerKey ed25519.PrivateKey
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
	shutdown := make(chan struct{}, 1)

	go func() {
		defer close(errs)
		defer func() {
			if err := ln.Close(); err != nil {
				errs <- err
			}
		}()

		for {
			select {
			case <-shutdown:
				// Graceful shutdown requested
				time.Sleep(100 * time.Millisecond) // Brief delay to ensure all messages are displayed
				fmt.Printf("%s\n", colorUtils.Success("Server shutdown complete."))
				return
			case <-ctx.Done():
				// Context cancelled
				return
			default:
				c, err := ln.Accept(ctx)
				if err != nil {
					// Accept returns error when listener is closed or ctx canceled.
					select {
					case errs <- err:
					default:
					}
					return
				}
				go func() {
					handleConn(ctx, c, s.ServerKey)
					// Signal shutdown after successful transfer
					select {
					case shutdown <- struct{}{}:
					default:
					}
				}()
			}
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
		fmt.Printf("%s %s\n", colorUtils.Info("Verify code:"), colorUtils.Highlight(FormatSAS(words)))
	}

	// 4) Receive Offer (list of files)
	var offer OfferMsg
	if err := ReadJSON(ctx, ctrl, &offer); err != nil {
		_ = c.CloseWithError(0, "bad offer")
		return
	}
	fmt.Printf("%s (%d):\n", colorUtils.Info("Incoming files"), len(offer.Files))
	for _, fm := range offer.Files {
		fmt.Printf("  %s %s %s\n", colorUtils.Seperator("•"), colorUtils.Bold(fm.Rel), colorUtils.Highlight(fmt.Sprintf("(%d bytes)", fm.Size)))
	}

	// 5) Destination directory: ~/.gns3
	home, _ := os.UserHomeDir()
	dst := filepath.Join(home, ".gns3")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		_ = c.CloseWithError(0, "mkdir failed")
		return
	}

	// 6) Send acceptance response to client
	if err := WriteJSON(ctx, ctrl, map[string]string{"status": "accepted"}); err != nil {
		_ = c.CloseWithError(0, "accept response failed")
		return
	}

	// 7) Receive the advertised number of files via unidirectional streams
	if err := ReceiveFiles(ctx, c, dst, len(offer.Files)); err != nil {
		_ = c.CloseWithError(0, "recv failed")
		return
	}
	fmt.Printf("%s\n", colorUtils.Success("Receive complete."))

	// 8) Send completion confirmation to client
	if err := WriteJSON(ctx, ctrl, map[string]string{"status": "complete"}); err != nil {
		_ = c.CloseWithError(0, "completion response failed")
		return
	}

	// 9) Wait for client to acknowledge completion before closing
	var ack map[string]string
	if err := ReadJSON(ctx, ctrl, &ack); err == nil {
		fmt.Printf("%s\n", colorUtils.Success("Transfer acknowledged by client"))
	}

	// 10) Send shutdown signal to main server loop
	fmt.Printf("%s\n", colorUtils.Info("Transfer complete. Server shutting down..."))

	// 11) Close gracefully
	_ = c.CloseWithError(0, "transfer complete")
}
