{ pkgs, ... }:

let
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
      if ${pkgs.iproute2}/bin/ip link show eth0 >/dev/null 2>&1; then
        break
      fi
      sleep 0.1
    done

    ${pkgs.iproute2}/bin/ip link set dev eth0 up
    ${pkgs.iproute2}/bin/ip addr flush dev eth0
    ${pkgs.iproute2}/bin/ip addr add "$runner_ipv4" dev eth0
    ${pkgs.iproute2}/bin/ip route replace default via "$runner_gw" dev eth0
  '';
in
{
  networking.useDHCP = false;
  networking.useNetworkd = false;
  networking.firewall.enable = false;

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
