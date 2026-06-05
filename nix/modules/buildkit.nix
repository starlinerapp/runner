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
    git
    curl
    cacert
  ];

  systemd.services.buildkitd = {
    description = "BuildKit Daemon";

    wantedBy = ["multi-user.target"];

    after = ["network-online.target"];
    wants = ["network-online.target"];

    serviceConfig = {
        Type = "simple";

        ExecStart = ''
          ${pkgs.buildkit}/bin/buildkitd \
            --addr unix:///run/buildkit/buildkitd.sock \
            --addr tcp://0.0.0.0:1234
        '';

        Restart = "always";
        RestartSec = 5;

        LimitNOFILE = 1048576;

        StateDirectory = "buildkit";
        RuntimeDirectory = "buildkit";
    };
  };
}