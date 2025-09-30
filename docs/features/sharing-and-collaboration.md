# Sharing and Collaboration

The `gns3util share` command group packages configuration artifacts so that administrators on the same network can exchange ready-to-use lab environments. Transfers rely on a secure QUIC tunnel with short-authentication-string (SAS) verification to prevent tampering by students.

## When to Use Sharing

- Hand off a fully provisioned cluster configuration to another instructor
- Sync updates to the GNS3 database (`clusterData.db`) across lab administrators

## Sending Artifacts

Run `share send` from the machine that already has the desired artifacts.

```bash
# Launch share send and choose artifacts/receiver interactively
gns3util share send

# Send cluster config, database, and key file to a discovered receiver
gns3util share send --all

# Send only the cluster config to another device discovered via mDNS
gns3util share send \
  --send-config \
  --src-dir ~/.gns3 \
  --discover-timeout 10

# Discover nearby receivers, fuzzy-pick, and send only the key file
gns3util share send \
  --send-key \
  --discover-timeout 5

# Share only the database file during a fuzzy picker window
gns3util share send \
  --send-db \
  --src-dir ~/.gns3
```

### Flags

- `--all`: Include `cluster_config.toml`, `clusterData.db`, and `gns3key`
- `--send-config`: Only send the config file
- `--send-db`: Only send the database file
- `--send-key`: Only send the key file
- `--discover-timeout <duration>`: Adjust the mDNS discovery window (default `3s`)
- `--src-dir <path>`: Choose a different artifact directory (default `~/.gns3`)
- `--to <label|host:port>`: Skip interactive selection and dial a specific receiver
- `--yes`: Accept prompts automatically (useful for automation)

If you omit `--to`, the command discovers nearby receivers via mDNS and lets you pick one interactively.

## Receiving Artifacts

Run `share receive` on the target machine and leave the terminal open while waiting for a sender.

```bash
# Host a receiver and wait for transfers
gns3util share receive
```

The command announces itself over mDNS, displays the SAS when a sender connects, and waits for you to confirm the transfer. Files are placed in your `~/.gns3` directory by default.

## Secure Workflow

1. **Receiver prepares**: Run `gns3util share receive` on the destination host.
2. **Sender selects artifacts**: Choose `--all` or the individual `--send-*` flags.
3. **Verify SAS**: Both sides compare the SAS displayed in their terminals.
4. **Complete transfer**: Approve on both ends; artifacts are copied over QUIC.
5. **Apply artifacts**: On the receiver, use `gns3util cluster config apply` or other commands to load the new data.

## Tips

- Use `--discover-timeout 0` to disable discovery if you must specify `--to` manually.
- Combine `--yes` with cron jobs or systemd timers for unattended nightly transfers.
- Combine sharing with version-controlled config files to keep lab environments reproducible.
- Run `gns3util share send --help` and `gns3util share receive --help` to view the full flag set and defaults before scripting automation.
- The receiver binds to a random UDP port and advertises it via mDNS; expect port numbers to change on each run. If you see a buffer warning from QUIC (similar to `failed to sufficiently increase receive buffer size`), follow the guidance at <https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes> to tune your OS settings.
