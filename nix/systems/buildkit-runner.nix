{ modulesPath, ... }:

{
  imports = [
    "${modulesPath}/profiles/minimal.nix"
    ../modules/buildkit.nix
    ../modules/networking.nix
    ../modules/firecracker.nix
  ];

  system.stateVersion = "26.05";

  boot.loader.grub.enable = false;

  boot.initrd.kernelModules = [
    "virtio_mmio"
    "virtio_blk"
    "virtio_net"
  ];

  fileSystems."/" = {
    device = "/dev/vda";
    fsType = "ext4";
  };

  boot.kernelParams = [
    "console=ttyS0"
    "reboot=k"
    "panic=1"
    "pci=off"
    "root=/dev/vda"
  ];

  services.openssh.enable = true;

  users.users.root = {
    initialPassword = "root";
  };
}
