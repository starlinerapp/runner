{
  description = "Starliner Runner";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    let
      linuxSystem = "x86_64-linux";

      buildkitRunner = nixpkgs.lib.nixosSystem {
        system = linuxSystem;

        modules = [
          ./nix/systems/buildkit-runner.nix
        ];
      };

      bootArgs = nixpkgs.lib.concatStringsSep " " (
        buildkitRunner.config.boot.kernelParams ++ [
          "init=${buildkitRunner.config.system.build.toplevel}/init"
        ]
      );

      linuxPackages =
        let
          kernel = buildkitRunner.config.boot.kernelPackages.kernel;
          initrd = buildkitRunner.config.system.build.initialRamdisk;
          initrdFile = buildkitRunner.config.system.boot.loader.initrdFile;
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

          buildkit-runner-bootargs = buildkitRunner.pkgs.runCommand "buildkit-runner-bootargs" { } ''
            mkdir -p $out
            cp ${buildkitRunner.pkgs.writeText "boot.args" bootArgs} $out/boot.args
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
            additionalSpace = "4G";
            copyChannel = false;
            name = "buildkit-runner-rootfs";
            baseName = "rootfs";
          };
        };
    in
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        darwinDeps = pkgs.lib.optionals pkgs.stdenv.isDarwin [
          pkgs.apple-sdk
          (pkgs.darwinMinVersionHook "11.0")
        ];

        commonPkgs = with pkgs; [
          git
          curl
          jq
          gnumake
          zstd
        ];

        devPkgs =
          with pkgs;
          [
            go
            golangci-lint
          ]
          ++ darwinDeps;

        shellHook = ''
          unset GOROOT
          export GOPATH="''${GOPATH:-$HOME/go}"
          export GOBIN="$GOPATH/bin"
          export PATH="$PATH:$GOBIN"
        '';

      in
      {
        formatter = pkgs.nixfmt-tree;
        devShells = {
          default = pkgs.mkShell {
            name = "starliner-runner";
            packages = commonPkgs ++ devPkgs;
            inherit shellHook;
          };
        };
      }
    )
    // {
      nixosConfigurations.buildkit-runner = buildkitRunner;
      packages.${linuxSystem} = linuxPackages;
    };
}
