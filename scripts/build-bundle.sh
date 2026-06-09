#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export NIX_CONFIG="extra-experimental-features = nix-command flakes ${NIX_CONFIG:-}"

RELEASE_TAG="${RELEASE_TAG:-dev}"
bundle="runner-${RELEASE_TAG}-linux-amd64"
bundle_dir="dist/${bundle}"

echo "==> Building Firecracker assets"
nix build .#buildkit-runner-rootfs --out-link result-rootfs
nix build .#buildkit-runner-vmlinux --out-link result-kernel
nix build .#buildkit-runner-initrd --out-link result-initrd
nix build .#buildkit-runner-bootargs --out-link result-bootargs

echo "==> Building runner CLI"
mkdir -p build
(
	cd cli
	go build -o ../build/runner ./cmd/main.go
)

echo "==> Packaging release bundle"
rm -rf "$bundle_dir"
mkdir -p "$bundle_dir"

cp build/runner "$bundle_dir/runner"
chmod +x "$bundle_dir/runner"

cp -L result-kernel/vmlinux "$bundle_dir/vmlinux"
cp -L result-initrd/initrd "$bundle_dir/initrd"
cp -L result-bootargs/boot.args "$bundle_dir/boot.args"
zstd -T0 -19 -f -o "$bundle_dir/rootfs.ext4.zst" result-rootfs/rootfs.img

rm -f "dist/${bundle}.tar"
tar -C dist -cf "dist/${bundle}.tar" "$bundle"

echo "==> Cleaning up build artifacts"
rm -f result-rootfs result-kernel result-initrd result-bootargs
rm -rf build

echo "==> Done: dist/${bundle}.tar"
