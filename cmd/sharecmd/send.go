package sharecmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/stefanistkuhl/gns3util/pkg/sharing/keys"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/mdns"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/transport"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/trust"
	pathUtils "github.com/stefanistkuhl/gns3util/pkg/utils/pathUtils"
)

// promptTrustCLI shows SAS code and asks to pin on first contact.
func promptTrustCLI(peerLabel, fp string, words []string) (bool, error) {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("First-time connection to %q\n", peerLabel)
	fmt.Printf("Fingerprint: %s\n", keys.ShortFingerprint(fp))
	fmt.Printf("Verify code: %s\n", transport.FormatSAS(words))
	fmt.Print("Do the codes match? Trust this device? [y/N]: ")
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

// selectReceiver discovers receivers via mDNS and returns a host:port to dial.
// If hint contains ":", returns it directly. If hint is a label, pre-filter.
// If hint empty, show menu of all.
func selectReceiver(ctx context.Context, hint string, timeout time.Duration) (addr string, label string, err error) {
	if hint != "" && strings.Contains(hint, ":") {
		return hint, hint, nil
	}

	fmt.Println("Discovering receivers on the LAN...")
	peers, err := mdns.Browse(ctx, timeout)
	if err != nil {
		return "", "", err
	}
	if len(peers) == 0 {
		return "", "", errors.New("no receivers found via mDNS; ensure the receiver is running and on the same LAN")
	}

	list := peers
	if hint != "" {
		filtered := make([]mdns.Peer, 0, len(peers))
		for _, p := range peers {
			if strings.EqualFold(p.Instance, hint) || strings.EqualFold(p.TXT["user"], hint) {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) == 1 {
			sel := filtered[0]
			if sel.Addr == "" {
				return "", "", fmt.Errorf("peer %q has no resolvable address", sel.Instance)
			}
			return sel.Addr, sel.Instance, nil
		}
		if len(filtered) > 1 {
			list = filtered
		} else {
			fmt.Printf("No exact match for %q; showing all discovered receivers.\n", hint)
		}
	}

	fmt.Println("Select a receiver:")
	for i, p := range list {
		fp := p.TXT["fp"]
		fmt.Printf("  [%d] %-30s %-22s fp=%s\n", i+1, p.Instance, p.Addr, keys.ShortFingerprint(fp))
	}
	fmt.Print("Enter number: ")
	var idx int
	_, scanErr := fmt.Scanf("%d\n", &idx)
	if scanErr != nil || idx < 1 || idx > len(list) {
		return "", "", errors.New("invalid selection")
	}
	chosen := list[idx-1]
	if chosen.Addr == "" {
		return "", "", fmt.Errorf("selected peer %q has no resolvable address", chosen.Instance)
	}
	return chosen.Addr, chosen.Instance, nil
}

func NewSendCmd() *cobra.Command {
	var (
		to              string
		discoverTimeout time.Duration
		srcDirFlag      string
		sendConfigFlag  bool
		sendDBFlag      bool
		sendKeyFlag     bool
		allFlag         bool
		yesFlag         bool
	)

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send GNS3 artifacts to a peer",
		Long:  "Discover or resolve a receiver, dial over QUIC, verify via SAS, pin on first contact, and transfer selected artifacts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load/create device key
			dk, err := keys.LoadOrCreate(keys.Options{})
			if err != nil {
				return err
			}
			fmt.Println("My device:", keys.DeviceLabel())
			fmt.Println("My FP:     ", keys.ShortFingerprint(dk.FP))

			// Trust store
			appDir, err := pathUtils.GetGNS3Dir()
			if err != nil {
				return err
			}
			ts, err := trust.Open(appDir)
			if err != nil {
				return err
			}

			// Source dir
			srcDir := srcDirFlag
			if srcDir == "" {
				home, _ := os.UserHomeDir()
				srcDir = filepath.Join(home, ".gns3")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Resolve target via mDNS selection if needed
			target, label, err := selectReceiver(ctx, to, discoverTimeout)
			if err != nil {
				return err
			}
			fmt.Printf("Dialing %s (%s)...\n", label, target)

			// Dial with SAS + pinning
			conn, ctrl, hello, err := transport.DialWithPin(
				ctx,
				target,
				keys.DeviceLabel(),
				dk.FP,
				dk.Pub,
				ts,
				promptTrustCLI,
			)
			if err != nil {
				return err
			}
			defer conn.CloseWithError(0, "done")

			fmt.Printf("Connected to %s as %q\n", target, hello.Label)

			// Determine which artifacts to send
			candidates := []string{"cluster_config.toml", "clusterData.db", "gns3key"}
			exists := func(rel string) (string, os.FileInfo, bool) {
				abs := filepath.Join(srcDir, rel)
				st, err := os.Stat(abs)
				if err == nil && st.Mode().IsRegular() {
					return abs, st, true
				}
				return abs, nil, false
			}

			selected := make([]string, 0, len(candidates))
			// Non-interactive flags take precedence
			if allFlag || sendConfigFlag || sendDBFlag || sendKeyFlag {
				want := map[string]bool{
					"cluster_config.toml": allFlag || sendConfigFlag,
					"clusterData.db":      allFlag || sendDBFlag,
					"gns3key":             allFlag || sendKeyFlag,
				}
				for rel, ok := range want {
					if !ok {
						continue
					}
					abs, st, ok := exists(rel)
					if ok {
						fmt.Printf("Include %s (%d bytes)\n", rel, st.Size())
						selected = append(selected, abs)
					} else {
						fmt.Printf("Skip %s (not found at %s)\n", rel, abs)
					}
				}
				if len(selected) == 0 {
					return errors.New("no artifacts selected or found")
				}
			} else {
				// Interactive per-file confirmation
				reader := bufio.NewReader(os.Stdin)
				for _, rel := range candidates {
					abs, st, ok := exists(rel)
					if !ok {
						continue
					}
					if yesFlag {
						fmt.Printf("Include %s (%d bytes)\n", rel, st.Size())
						selected = append(selected, abs)
						continue
					}
					fmt.Printf("Send %s (%d bytes)? [y/N]: ", rel, st.Size())
					line, _ := reader.ReadString('\n')
					line = strings.TrimSpace(strings.ToLower(line))
					if line == "y" || line == "yes" {
						selected = append(selected, abs)
					}
				}
				if len(selected) == 0 {
					fmt.Println("Nothing selected; aborting.")
					return nil
				}
				if !yesFlag {
					// Final confirmation
					fmt.Println("About to send:")
					for _, abs := range selected {
						rel := filepath.Base(abs)
						st, _ := os.Stat(abs)
						fmt.Printf("  - %s (%d bytes)\n", rel, st.Size())
					}
					fmt.Print("Proceed? [y/N]: ")
					line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
					line = strings.TrimSpace(strings.ToLower(line))
					if line != "y" && line != "yes" {
						fmt.Println("Aborted.")
						return nil
					}
				}
			}

			// Send offer + files
			if err := transport.SendOfferAndFiles(ctx, ctrl, conn, selected); err != nil {
				return err
			}
			fmt.Println("Send complete.")
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "receiver label or host:port (omit to pick from a list)")
	cmd.Flags().DurationVar(&discoverTimeout, "discover-timeout", 3*time.Second, "mDNS discovery window")
	cmd.Flags().StringVar(&srcDirFlag, "src-dir", "", "source directory for artifacts (default: ~/.gns3)")
	cmd.Flags().BoolVar(&allFlag, "all", false, "send all artifacts (config, db, key)")
	cmd.Flags().BoolVar(&sendConfigFlag, "send-config", false, "include cluster_config.toml")
	cmd.Flags().BoolVar(&sendDBFlag, "send-db", false, "include clusterData.db")
	cmd.Flags().BoolVar(&sendKeyFlag, "send-key", false, "include gns3key")
	cmd.Flags().BoolVar(&yesFlag, "yes", false, "assume yes for all prompts (non-interactive)")
	return cmd
}
