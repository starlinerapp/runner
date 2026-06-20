# Starliner Runner

Build and publish container images in Firecracker microVMs on Linux.

## Requirements

- Linux (amd64)
- `sudo` for install

## Quick start

```bash
tar xf runner-linux-amd64.tar
cd runner-linux-amd64
sudo ./runner install
sudo ./runner register
sudo ./runner start
```

Run `install` once from the extracted bundle. Register the runner with Starliner, then start the daemon to claim and execute remote build jobs.

## Commands

| Command | Description |
|---------|-------------|
| `runner install` | Install the CLI, VM assets, Firecracker, and BuildKit |
| `runner register` | Register this runner with Starliner |
| `runner start` | Start the runner daemon |
| `runner vm create` | Create a microVM |
| `runner vm list` | List microVMs |
| `runner vm delete <id>` | Remove a microVM |

## Development

```bash
./scripts/build-bundle.sh
cd dist/runner-dev-linux-amd64
sudo ./runner install
```

Use `./scripts/build-bundle.sh --cli-only` to rebuild the CLI without rebuilding the guest image.
