# Starliner Runner

Build and publish container images in Firecracker microVMs on Linux.

## Requirements

- Linux (amd64)
- `sudo` for install and builds

## Quick start

```bash
tar xf runner-linux-amd64.tar
cd runner-linux-amd64
sudo ./runner install

sudo runner build \
  --repository https://github.com/org/app.git \
  --branch-name main \
  --github-token "$GITHUB_TOKEN" \
  --image registry.example.com/app \
  --registry-username "$REGISTRY_USER" \
  --registry-password "$REGISTRY_PASSWORD"
```

Run `install` once from the extracted bundle. Builds use the installed binary at `/usr/local/bin/runner`. Images are tagged with the cloned commit SHA.

## Commands

| Command | Description |
|---------|-------------|
| `runner install` | Install the CLI, VM assets, Firecracker, and BuildKit |
| `runner build` | Clone a repo, build a Dockerfile in a VM, and push to a registry |
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
