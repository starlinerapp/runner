{ pkgs, ... }:

{
  boot.kernelModules = [
    "overlay"
    "virtio_net"
  ];

  boot.supportedFilesystems = [
    "overlay"
  ];

  environment.systemPackages = with pkgs; [
    buildkit
    runc
    git
    curl
    cacert
  ];

  environment.etc."buildkit/buildkitd.toml".text = ''
    [worker.oci]
      enabled = true
      binary = "${pkgs.runc}/bin/runc"
      rootless = false

    [worker.containerd]
      enabled = false
  '';

  systemd.services.buildkitd = {
    description = "BuildKit Daemon";

    wantedBy = [ "multi-user.target" ];

    after = [ "network.target" ];

    path = with pkgs; [
      buildkit
      runc
      iptables
    ];

    serviceConfig = {
      Type = "simple";

      ExecStart = ''
        ${pkgs.buildkit}/bin/buildkitd \
          --config /etc/buildkit/buildkitd.toml \
          --addr unix:///run/buildkit/buildkitd.sock \
          --addr tcp://0.0.0.0:1234
      '';

      Restart = "always";
      RestartSec = 5;

      LimitNOFILE = 1048576;

      StateDirectory = "buildkit";
      RuntimeDirectory = "buildkit";

      Delegate = true;
    };
  };
}
