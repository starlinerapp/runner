{
    description = "Starliner Runner";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    };

    outputs = { self, nixpkgs }:
    let
        system = "x86_64-linux";

        buildkitRunner = nixpkgs.lib.nixosSystem {
            inherit system;

            modules = [
                ./nix/images/buildkit-runner.nix
            ];
        };
    in
    {
        nixosConfigurations.buildkit-runner = buildkitRunner;

        packages.${system} =
            let
                kernel = buildkitRunner.config.boot.kernelPackages.kernel;
            in
            {
                buildkit-runner-kernel = kernel.dev;

                buildkit-runner-vmlinux = buildkitRunner.pkgs.runCommand "buildkit-runner-vmlinux" { } ''
                    mkdir -p $out
                    cp ${kernel.dev}/vmlinux $out/vmlinux
                '';

                buildkit-runner-rootfs = import "${nixpkgs}/nixos/lib/make-disk-image.nix" {
                pkgs = buildkitRunner.pkgs;
                inherit (buildkitRunner.pkgs) lib;
                config = buildkitRunner.config;
                format = "raw";
                fsType = "ext4";
                partitionTableType = "none";
                installBootLoader = false;
                diskSize = "auto";
                additionalSpace = "512M";
                copyChannel = false;
                name = "buildkit-runner-rootfs";
                baseName = "rootfs";
            };
        };
    };
}
