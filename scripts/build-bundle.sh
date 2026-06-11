#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

CLI_ONLY=false

while [[ $# -gt 0 ]]; do
	case "$1" in
		--cli-only)
			CLI_ONLY=true
			shift
			;;
		*)
			echo "Unknown argument: $1" >&2
			exit 1
			;;
	esac
done

export NIX_CONFIG="extra-experimental-features = nix-command flakes ${NIX_CONFIG:-}"

RELEASE_TAG="${RELEASE_TAG:-dev}"
bundle="runner-${RELEASE_TAG}-linux-amd64"
bundle_dir="dist/${bundle}"

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

if [[ "$CLI_ONLY" != "true" ]]; then
	echo "==> Building Firecracker assets"
	nix build .#buildkit-runner-rootfs --out-link result-rootfs
	nix build .#buildkit-runner-vmlinux --out-link result-kernel
	nix build .#buildkit-runner-initrd --out-link result-initrd
	nix build .#buildkit-runner-bootargs --out-link result-bootargs

	cp -L result-kernel/vmlinux "$bundle_dir/vmlinux"
	cp -L result-initrd/initrd "$bundle_dir/initrd"
	cp -L result-bootargs/boot.args "$bundle_dir/boot.args"
	zstd -T0 -19 -f -o "$bundle_dir/rootfs.ext4.zst" result-rootfs/rootfs.img
fi

rm -f "dist/${bundle}.tar"
tar -C dist -cf "dist/${bundle}.tar" "$bundle"

echo "==> Cleaning up build artifacts"
rm -f result-rootfs result-kernel result-initrd result-bootargs
rm -rf build

echo "==> Done: dist/${bundle}.tar"