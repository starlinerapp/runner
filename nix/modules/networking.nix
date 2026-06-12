{
  networking.useDHCP = false;
  networking.useNetworkd = true;

  networking.firewall.enable = false;

  systemd.network.enable = true;
  systemd.network.wait-online.enable = false;

  # Guest IP is assigned by kernel ip= boot arg; keep networkd from reconfiguring it.
  systemd.network.networks."10-eth0" = {
    matchConfig.Name = "eth0";
    networkConfig.KeepConfiguration = "static";
  };
}
