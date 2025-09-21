package sharecmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/stefanistkuhl/gns3util/pkg/sharing/keys"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/mdns"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/transport"
)

func NewReceiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "receive",
		Short: "Receive from a peer",
		Long:  "Start a QUIC listener, advertise via mDNS, show SAS on first contact, and wait for transfers",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1) Load/create device key
			dk, err := keys.LoadOrCreate(keys.Options{})
			if err != nil {
				return err
			}
			fmt.Println("My device:", keys.DeviceLabel())
			fmt.Println("My FP:     ", keys.ShortFingerprint(dk.FP))

			// 2) TLS config (with ALPN) from ed25519 key
			cert, err := transport.CertFromEd25519(dk.Priv, "gns3util")
			if err != nil {
				return err
			}
			tlsConf := transport.ServerTLS(cert)

			// 3) QUIC server with server key for SAS
			srv := transport.Server{
				TLS:       tlsConf,
				ServerKey: dk.Priv, // needed to derive SAS on server
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			addr, errs, closeFn, err := srv.Listen(ctx, ":0")
			if err != nil {
				return err
			}
			defer func() {
				if err := closeFn(); err != nil {
					log.Println("close listener:", err)
				}
			}()

			// 4) Print UDP address
			var port int
			if udp, ok := addr.(*net.UDPAddr); ok {
				port = udp.Port
				fmt.Printf("Listening on %s (UDP %d)\n", udp.IP.String(), udp.Port)
			} else {
				fmt.Printf("Listening on %v\n", addr)
			}

			// 5) mDNS advertise
			var stopAdv func()
			if port != 0 {
				adv := &mdns.Advertiser{}
				user := os.Getenv("USER")
				if user == "" {
					user = os.Getenv("USERNAME")
				}
				host, _ := os.Hostname()
				stopAdv, err = adv.Start(
					keys.DeviceLabel(), // instance: "user@host"
					port,
					map[string]string{
						"fp":   dk.FP,
						"ver":  "1",
						"user": user,
						"host": host,
					},
				)
				if err != nil {
					fmt.Println("mDNS advertise failed:", err)
				} else {
					defer stopAdv()
					fmt.Printf("Advertised via mDNS as %q (_gns3util-share._udp)\n", keys.DeviceLabel())
				}
			}

			fmt.Println("Waiting for a connection...")

			// 6) Wait for either an accept-loop error or timeout
			select {
			case err := <-errs:
				return err
			case <-time.After(5 * time.Minute):
				fmt.Println("No incoming connections in 5m; exiting")
				return nil
			}
		},
	}

	return cmd
}
