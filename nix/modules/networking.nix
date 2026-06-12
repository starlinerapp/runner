{
  networking.useDHCP = true;
  networking.useNetworkd = true;

  networking.firewall.enable = false;

  systemd.network.enable = true;
  systemd.network.wait-online.enable = true;
  systemd.network.wait-online.anyInterface = true;
}
