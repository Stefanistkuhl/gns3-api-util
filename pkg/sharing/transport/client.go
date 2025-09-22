package transport

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/quic-go/quic-go"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/keys"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/trust"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

type VerifyPrompt func(peerLabel, fp string, words []string) (bool, error)

func DialWithPin(
	ctx context.Context,
	addr string,
	myLabel string,
	myFP string,
	myPub ed25519.PublicKey,
	ts *trust.Store,
	prompt VerifyPrompt,
) (*quic.Conn, *quic.Stream, Hello, error) {
	tlsConf := &tls.Config{
		NextProtos:         []string{"gns3util/1"},
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
	}

	conn, err := quic.DialAddr(ctx, addr, tlsConf, &quic.Config{})
	if err != nil {
		return nil, nil, Hello{}, err
	}

	ctrl, err := conn.OpenStreamSync(ctx)
	if err != nil {
		_ = conn.CloseWithError(0, "open stream failed")
		return nil, nil, Hello{}, err
	}

	// 1) Hello
	if err := WriteJSON(ctx, ctrl, Hello{Label: myLabel, FP: myFP}); err != nil {
		_ = conn.CloseWithError(0, "hello failed")
		return nil, nil, Hello{}, err
	}
	var srv Hello
	if err := ReadJSON(ctx, ctrl, &srv); err != nil {
		_ = conn.CloseWithError(0, "hello recv failed")
		return nil, nil, Hello{}, err
	}

	// 2) SAS nonce exchange
	clientNonce, err := NewNonce()
	if err != nil {
		_ = conn.CloseWithError(0, "nonce gen failed")
		return nil, nil, Hello{}, err
	}
	if err := WriteJSON(ctx, ctrl, SASMsg{Nonce: clientNonce}); err != nil {
		_ = conn.CloseWithError(0, "sas write failed")
		return nil, nil, Hello{}, err
	}
	var srvSAS SASMsg
	if err := ReadJSON(ctx, ctrl, &srvSAS); err != nil {
		_ = conn.CloseWithError(0, "sas read failed")
		return nil, nil, Hello{}, err
	}

	// 3) Extract server public key from TLS to compute fingerprint and SAS
	st := conn.ConnectionState().TLS
	if len(st.PeerCertificates) == 0 {
		_ = conn.CloseWithError(0, "no peer cert")
		return nil, nil, Hello{}, errors.New("no peer certificate")
	}
	cert := st.PeerCertificates[0]
	serverPub, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		_ = conn.CloseWithError(0, "unexpected key type")
		return nil, nil, Hello{}, errors.New("server key is not ed25519")
	}
	serverFP := keys.Fingerprint(serverPub)

	// 4) Derive SAS code bound to server identity + fresh nonces
	words, err := DerivePGPWordsSimple(serverPub, clientNonce, srvSAS.Nonce, 3)
	if err != nil {
		_ = conn.CloseWithError(0, "sas derive failed")
		return nil, nil, Hello{}, err
	}
	fmt.Printf("%s %s\n", colorUtils.Info("Verify code:"), colorUtils.Highlight(FormatSAS(words)))

	// 5) Pinning: if not pinned, ask the user to accept
	if _, ok := ts.Get(serverFP); !ok {
		if prompt == nil {
			_ = conn.CloseWithError(0, "unpinned and no prompt")
			return nil, nil, Hello{}, errors.New("unpinned and no prompt")
		}
		accept, err := prompt(srv.Label, serverFP, words)
		if err != nil || !accept {
			_ = conn.CloseWithError(0, "verification rejected")
			if err == nil {
				err = errors.New("verification rejected")
			}
			return nil, nil, Hello{}, err
		}
		// Persist pin
		_ = ts.Add(serverFP, srv.Label)
	}

	return conn, ctrl, srv, nil
}
