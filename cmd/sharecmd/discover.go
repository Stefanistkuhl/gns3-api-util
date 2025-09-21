package sharecmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/keys"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/mdns"
)

func NewDiscoverCmd() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover peers on the LAN via mDNS",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			peers, err := mdns.Browse(ctx, timeout)
			if err != nil {
				return err
			}
			if len(peers) == 0 {
				fmt.Println("No peers found.")
				return nil
			}
			for _, p := range peers {
				fp := p.TXT["fp"]
				fmt.Printf("- %s @ %s  fp=%s\n", p.Instance, p.Addr, keys.ShortFingerprint(fp))
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "discovery time window")
	return cmd
}
