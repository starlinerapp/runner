#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RELEASE_TAG="${RELEASE_TAG:-dev}"
bundle="runner-${RELEASE_TAG}-linux-amd64"

echo "==> Building Firecracker assets"
nix build .#buildkit-runner-rootfs --out-link result-rootfs
nix build .#buildkit-runner-vmlinux --out-link result-kernel
nix build .#buildkit-runner-initrd --out-link result-initrd
nix build .#buildkit-runner-bootargs --out-link result-bootargs

echo "==> Building runner CLI"
(
	cd cli
	go build -o ../build/runner ./cmd/main.go
)

echo "==> Packaging release bundle"
mkdir -p "dist/${bundle}"

cp build/runner "dist/${bundle}/runner"
chmod +x "dist/${bundle}/runner"
cp -L result-kernel/vmlinux "dist/${bundle}/vmlinux"
cp -L result-initrd/initrd "dist/${bundle}/initrd"
cp -L result-bootargs/boot.args "dist/${bundle}/boot.args"
zstd -T0 -19 -f -o "dist/${bundle}/rootfs.ext4.zst" result-rootfs/rootfs.img

tar -C dist -cf "dist/${bundle}.tar" "${bundle}"

echo "==> Done: dist/${bundle}.tar"
