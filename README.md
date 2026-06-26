# Starliner Runner

Build and publish container images in Firecracker microVMs on Linux.

## Requirements

- Linux (amd64)
- `sudo` for install

## Quick start

```bash
tar xf runner-linux-amd64.tar
cd runner-linux-amd64
./runner install --baseUrl <baseUrl>
sudo runner register --token <token>
sudo systemctl enable --now starliner-runner
```

## Commands

| Command                                | Description |
|----------------------------------------|-------------|
| `./runner install --baseUrl <baseUrl>` | Install the CLI, VM assets, Firecracker, BuildKit, and systemd service |
| `sudo runner register --token <token>` | Register this runner with Starliner |
| `sudo runner start`                    | Start the runner in the foreground (for debugging) |
| `runner vm create`                     | Create a microVM |
| `runner vm list`                       | List microVMs |
| `runner vm delete <id>`                | Remove a microVM |

## Service

```bash
sudo systemctl status starliner-runner
sudo journalctl -u starliner-runner -f
sudo systemctl stop starliner-runner
sudo systemctl restart starliner-runner
```

### Skip TLS verification

```bash
sudo ./runner register --token ... --insecure-skip-tls-verify
```

```bash
sudo mkdir -p /etc/systemd/system/starliner-runner.service.d
sudo tee /etc/systemd/system/starliner-runner.service.d/insecure-tls.conf <<'EOF'
[Service]
ExecStart=
ExecStart=/usr/local/bin/runner start --insecure-skip-tls-verify
EOF

sudo systemctl daemon-reload
sudo systemctl restart starliner-runner
```

```bash
systemctl cat starliner-runner
```

## Development

```bash
./scripts/build-bundle.sh
cd dist/runner-dev-linux-amd64
sudo ./runner install
```

```bash
./scripts/build-bundle.sh --cli-only
```
