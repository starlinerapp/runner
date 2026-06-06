{
  networking.useDHCP = true;
  networking.useNetworkd = true;

  networking.firewall.enable = false;

  systemd.network.enable = true;
}
