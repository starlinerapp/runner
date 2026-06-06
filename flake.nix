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
                initrd = buildkitRunner.config.system.build.initialRamdisk;
                initrdFile = buildkitRunner.config.system.boot.loader.initrdFile;
                cfg = buildkitRunner.config;
                inherit (buildkitRunner.pkgs) lib;
            in
            {
                buildkit-runner-kernel = kernel.dev;

                buildkit-runner-vmlinux = buildkitRunner.pkgs.runCommand "buildkit-runner-vmlinux" { } ''
                    mkdir -p $out
                    cp ${kernel.dev}/vmlinux $out/vmlinux
                '';

                buildkit-runner-initrd = buildkitRunner.pkgs.runCommand "buildkit-runner-initrd" { } ''
                    mkdir -p $out
                    cp ${initrd}/${initrdFile} $out/initrd
                '';

                buildkit-runner-firecracker-config = buildkitRunner.pkgs.writeText "firecracker.json" (builtins.toJSON {
                    "boot-source" = {
                        kernel_image_path = "./vmlinux";
                        initrd_path = "./initrd";
                        boot_args = lib.concatStringsSep " " (
                            cfg.boot.kernelParams ++ [
                                "init=${cfg.system.build.toplevel}/init"
                            ]
                        );
                    };
                    drives = [
                        {
                            drive_id = "rootfs";
                            path_on_host = "./rootfs.ext4";
                            is_root_device = true;
                            is_read_only = false;
                        }
                    ];
                    network-interfaces = [
                        {
                            iface_id = "eth0";
                            guest_mac = "AA:FC:00:00:00:01";
                            host_dev_name = "tap0";
                        }
                    ];
                    "machine-config" = {
                        vcpu_count = 2;
                        mem_size_mib = 2048;
                    };
                });

                buildkit-runner-rootfs = import "${nixpkgs}/nixos/lib/make-disk-image.nix" {
                    pkgs = buildkitRunner.pkgs;
                    inherit (buildkitRunner.pkgs) lib;
                    config = buildkitRunner.config;
                    format = "raw";
                    fsType = "ext4";
                    partitionTableType = "none";
                    installBootLoader = false;
                    diskSize = "auto";
                    additionalSpace = "256M";
                    copyChannel = false;
                    name = "buildkit-runner-rootfs";
                    baseName = "rootfs";
                };
            };
    };
}
