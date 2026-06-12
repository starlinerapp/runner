{
  networking.useDHCP = false;
  networking.useNetworkd = false;
  networking.firewall.enable = false;

  systemd.network.wait-online.enable = false;
}
