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

	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/keys"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/mdns"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/transport"
	"github.com/stefanistkuhl/gns3util/pkg/sharing/trust"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	pathUtils "github.com/stefanistkuhl/gns3util/pkg/utils/pathUtils"
)

func promptTrustCLI(peerLabel, fp string, words []string) (bool, error) {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s %s\n", colorUtils.Warning("First-time connection to"), colorUtils.Bold(peerLabel))
	fmt.Printf("%s %s\n", colorUtils.Info("Fingerprint:"), colorUtils.Highlight(keys.ShortFingerprint(fp)))
	fmt.Printf("%s %s\n", colorUtils.Info("Verify code:"), colorUtils.Highlight(transport.FormatSAS(words)))
	fmt.Printf("%s ", colorUtils.Bold("Do the codes match? Trust this device? [y/N]:"))
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

func selectReceiver(ctx context.Context, hint string, timeout time.Duration) (addr string, label string, err error) {
	if hint != "" && strings.Contains(hint, ":") {
		return hint, hint, nil
	}

	fmt.Printf("%s\n", colorUtils.Info("Discovering receivers on the LAN..."))

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
			fmt.Printf("%s\n", colorUtils.Warning("No exact match for %q; showing all discovered receivers.", hint))
		}
	}

	options := make([]string, len(list))
	peerMap := make(map[string]mdns.Peer)

	for i, p := range list {
		fp := p.TXT["fp"]
		shortFP := ""
		if len(fp) >= 4 {
			shortFP = fp[:4]
		}
		option := fmt.Sprintf("%-30s │ %s", p.Instance, shortFP)
		options[i] = option
		peerMap[option] = p
	}

	selected := fuzzy.NewFuzzyFinderWithTitle(options, false, "Select a receiver:")
	if len(selected) == 0 {
		return "", "", errors.New("no receiver selected")
	}

	chosen, exists := peerMap[selected[0]]
	if !exists {
		return "", "", errors.New("invalid selection")
	}

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
			dk, err := keys.LoadOrCreate(keys.Options{})
			if err != nil {
				return err
			}
			fmt.Printf("%s %s\n", colorUtils.Info("My device:"), colorUtils.Bold(keys.DeviceLabel()))
			fmt.Printf("%s %s\n", colorUtils.Info("My FP:     "), colorUtils.Highlight(keys.ShortFingerprint(dk.FP)))

			appDir, err := pathUtils.GetGNS3Dir()
			if err != nil {
				return err
			}
			ts, err := trust.Open(appDir)
			if err != nil {
				return err
			}

			srcDir := srcDirFlag
			if srcDir == "" {
				home, _ := os.UserHomeDir()
				srcDir = filepath.Join(home, ".gns3")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			target, label, err := selectReceiver(ctx, to, discoverTimeout)
			if err != nil {
				return err
			}
			fmt.Printf("%s %s (%s)...\n", colorUtils.Info("Dialing"), colorUtils.Bold(label), colorUtils.Highlight(target))

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
			defer func() { _ = conn.CloseWithError(0, "done") }()

			fmt.Printf("%s %s as %s\n", colorUtils.Success("Connected to"), colorUtils.Highlight(target), colorUtils.Bold(fmt.Sprintf("\"%s\"", hello.Label)))

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
						fmt.Printf("%s %s %s\n", colorUtils.Success("Include"), colorUtils.Bold(rel), colorUtils.Highlight(fmt.Sprintf("(%d bytes)", st.Size())))
						selected = append(selected, abs)
					} else {
						fmt.Printf("%s %s %s\n", colorUtils.Warning("Skip"), colorUtils.Bold(rel), colorUtils.Seperator(fmt.Sprintf("(not found at %s)", abs)))
					}
				}
				if len(selected) == 0 {
					return errors.New("no artifacts selected or found")
				}
			} else {
				availableFiles := make([]string, 0)
				fileMap := make(map[string]string)

				for _, rel := range candidates {
					abs, st, ok := exists(rel)
					if !ok {
						continue
					}
					plainName := fmt.Sprintf("%-20s (%d bytes)", rel, st.Size())
					availableFiles = append(availableFiles, plainName)
					fileMap[plainName] = abs
				}

				if len(availableFiles) == 0 {
					return errors.New("no artifacts found in source directory")
				}

				if yesFlag {
					for _, abs := range fileMap {
						selected = append(selected, abs)
					}
				} else {
					selectedFiles := fuzzy.NewFuzzyFinderWithTitle(availableFiles, true, "Select files to send:")

					if len(selectedFiles) == 0 {
						fmt.Printf("%s\n", colorUtils.Warning("Nothing selected; aborting."))
						return nil
					}

					for _, displayName := range selectedFiles {
						if abs, ok := fileMap[displayName]; ok {
							selected = append(selected, abs)
						}
					}
				}

				fmt.Printf("\n%s\n", colorUtils.Info("About to send:"))
				for _, abs := range selected {
					rel := filepath.Base(abs)
					st, _ := os.Stat(abs)
					fmt.Printf("  %s %s %s\n", colorUtils.Seperator("•"), colorUtils.Bold(rel), colorUtils.Highlight(fmt.Sprintf("(%d bytes)", st.Size())))
				}
			}

			if err := transport.SendOfferAndFiles(ctx, ctrl, conn, selected); err != nil {
				return err
			}
			fmt.Printf("%s\n", colorUtils.Success("Send complete."))
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
