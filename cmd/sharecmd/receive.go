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
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
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
			fmt.Printf("%s %s\n", colorUtils.Info("My device:"), colorUtils.Bold(keys.DeviceLabel()))
			fmt.Printf("%s %s\n", colorUtils.Info("My FP:     "), colorUtils.Highlight(keys.ShortFingerprint(dk.FP)))

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
				fmt.Printf("%s %s %s\n", colorUtils.Success("Listening on"), colorUtils.Highlight(udp.IP.String()), colorUtils.Info(fmt.Sprintf("(UDP %d)", udp.Port)))
			} else {
				fmt.Printf("%s %v\n", colorUtils.Success("Listening on"), addr)
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
					fmt.Printf("%s %v\n", colorUtils.Warning("mDNS advertise failed:"), err)
				} else {
					defer stopAdv()
					fmt.Printf("%s %s %s\n", colorUtils.Success("Advertised via mDNS as"), colorUtils.Bold(keys.DeviceLabel()), colorUtils.Seperator("(_gns3util-share._udp)"))
				}
			}

			fmt.Printf("%s\n", colorUtils.Info("Waiting for a connection..."))

			// 6) Wait for either an accept-loop error or timeout
			select {
			case err := <-errs:
				if err != nil {
					return err
				}
				// err == nil means graceful shutdown after successful transfer
				return nil
			case <-time.After(5 * time.Minute):
				fmt.Printf("%s\n", colorUtils.Warning("No incoming connections in 5m; exiting"))
				return nil
			}
		},
	}

	return cmd
}
