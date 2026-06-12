{ pkgs, ... }:

let
  ip = "${pkgs.iproute2}/bin/ip";

  configureGuestNet = pkgs.writeShellScript "runner-configure-guest-net" ''
    set -euo pipefail

    runner_ipv4=$(
      tr ' ' '\n' </proc/cmdline |
        sed -n 's/^runner\.ipv4=//p' |
        head -1
    )
    runner_gw=$(
      tr ' ' '\n' </proc/cmdline |
        sed -n 's/^runner\.ipv4_gw=//p' |
        head -1
    )

    if [ -z "$runner_ipv4" ] || [ -z "$runner_gw" ]; then
      echo "runner.ipv4 and runner.ipv4_gw kernel parameters are required" >/dev/console
      exit 1
    fi

    for _ in $(seq 1 100); do
      if ${ip} link show eth0 >/dev/null 2>&1; then
        break
      fi
      sleep 0.1
    done

    ${ip} link set dev eth0 up
    ${ip} addr flush dev eth0
    ${ip} addr add "$runner_ipv4" dev eth0
    ${ip} route replace default via "$runner_gw" dev eth0

    # Static IP does not use host dnsmasq; set DNS
    printf 'nameserver 8.8.8.8\nnameserver 1.1.1.1\n' >/etc/resolv.conf
  '';
in
{
  networking.useDHCP = false;
  networking.useNetworkd = false;
  networking.firewall.enable = false;
  networking.nameservers = [ "8.8.8.8" "1.1.1.1" ];

  systemd.network.wait-online.enable = false;

  systemd.services.runner-guest-net = {
    description = "Configure static guest IP from kernel cmdline";

    wantedBy = [ "multi-user.target" ];
    before = [ "buildkitd.service" ];
    after = [ "systemd-udev-settle.service" ];

    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = true;
      ExecStart = configureGuestNet;
    };
  };
}
