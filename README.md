# Starliner Runner

The runner is the application that runs builds in Starliner. It provisions Firecracker microVMs on a Linux host, each with its own network, disk, and kernel.

## Requirements

- Linux (amd64)
- `sudo` access for install and VM networking

## Quick start

```bash
curl -L -o runner-v0.0.1-linux-amd64.tar \
  https://github.com/starlinerapp/runner/releases/download/v0.0.1/runner-v0.0.1-linux-amd64.tar

tar xf runner-v0.0.1-linux-amd64.tar
cd runner-v0.0.1-linux-amd64
./runner install

runner vm create
runner vm list
runner vm delete <id>
```

`./runner install` is a one-time step run from the extracted bundle. It installs the CLI, VM assets, and Firecracker.

## Commands

| Command | Description |
|---------|-------------|
| `runner install` | Install the CLI, assets, and Firecracker to the host |
| `runner vm create` | Provision a new microVM |
| `runner vm list` | List running VMs |
| `runner vm delete <id>` | Stop and remove a VM |

## Architecture

### Release bundle

| File | Purpose |
|------|---------|
| `runner` | CLI binary |
| `vmlinux` | Linux kernel |
| `initrd` | Initramfs |
| `boot.args` | Kernel command line matched to the rootfs |
| `rootfs.ext4.zst` | Compressed root disk image |

### On-disk layout

After install:

| Path | Contents |
|------|----------|
| `/usr/local/bin/runner` | CLI binary |
| `/usr/local/bin/firecracker` | Firecracker binary (installed if missing) |
| `/usr/local/share/runner/` | `vmlinux`, `initrd`, `boot.args`, decompressed `rootfs.ext4` |

At runtime:

| Path | Contents |
|------|----------|
| `~/.cache/runner/vms.json` | VM registry |
| `~/.cache/runner/vms/<id>/` | Per-VM workspace (rootfs copy, config, logs) |

Install decompresses `rootfs.ext4.zst` into `/usr/local/share/runner/`. Each `vm create` copies that image into the VM workspace so VMs are isolated.

### VM lifecycle

**Create** — `runner vm create` does the following:

1. Allocates a VM id, tap device (`rtap-<id>`), MAC, guest CID, and subnet (`172.16.<n>.0/24`)
2. Prepares a workspace under `~/.cache/runner/vms/<id>/` with a rootfs copy, kernel/initrd symlinks, and a `firecracker.json` config (2 vCPUs, 2 GiB RAM)
3. Creates a TAP device, enables NAT, and starts `dnsmasq` for guest DHCP
4. Starts Firecracker and records the VM in `vms.json`

Up to 254 VMs can run concurrently on one host.

**List** — reads `vms.json` and prints each VM's id, network details, workspace, log path, and Firecracker PID.

**Delete** — stops Firecracker, tears down the TAP device and `dnsmasq`, removes the workspace, and deletes the registry entry.

### Networking

Each VM gets a private `172.16.<n>.0/24` network. The host is the gateway (`172.16.<n>.1`), assigns addresses via DHCP, and NATs guest traffic through the host's default route.
