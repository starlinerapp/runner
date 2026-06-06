{ config, ... }:

{
  system.activationScripts.runnerInit.text = ''
    ln -sfn ${config.system.build.toplevel}/init /init
  '';
}
