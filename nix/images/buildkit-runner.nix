{ modulesPath, ... }:

{
  imports = [
    "${modulesPath}/profiles/minimal.nix"
    ../modules/buildkit.nix
    ../modules/networking.nix
  ];

  system.stateVersion = "26.05";

  boot.loader.grub.enable = false;

  boot.initrd.kernelModules = [
    "virtio_mmio"
    "virtio_blk"
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
    "init=${config.system.build.toplevel}/init"
  ];

  services.openssh.enable = true;

  users.users.root = {
    initialPassword = "root";
  };
}