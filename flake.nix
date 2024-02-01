{
  description = "Kurtosis dev flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
  };

  outputs = { self, nixpkgs, unstable, flake-utils, gomod2nix, ... }:
    let utils = flake-utils;
    in utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        unstable_pkgs = unstable.legacyPackages.${system};
        rev = "${self.shortRev or self.dirtyRev or "dirty"}";
      in rec {
        formatter = pkgs.nixpkgs-fmt;

        devShells.default = pkgs.callPackage ./shell.nix {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        };

        packages.default = packages.cli;

        packages.cli = pkgs.callPackage ./cli/cli/. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.engine = pkgs.callPackage ./engine/server/. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.enclave-manager = pkgs.callPackage ./enclave-manager/server/. {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.core = pkgs.callPackage ./core/server/. {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.files_artifacts_expander =
          pkgs.callPackage ./core/files_artifacts_expander/. {
            inherit rev;
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

        checks.cli = packages.cli;
        checks.core = packages.core;
        checks.engine = packages.engine;

        container.image.amd64 = let
          server = packages.default.overrideAttrs (old:
            old // {
              GOOS = "linux";
              GOARCH = "amd64";
              doCheck = false;
            });
        in pkgs.dockerTools.buildImage {
          name = "kurtosis-cloud-backend";
          tag = rev;
          created = "now";
          contents = server;
          architecture = "amd64";
          config.Cmd = [ "${server}/bin/linux_amd64/server" ];
        };

        container.image.arm64 = let
          server = packages.default.overrideAttrs (old:
            old // {
              GOOS = "linux";
              GOARCH = "arm64";
              doCheck = false;
            });
        in pkgs.dockerTools.buildImage {
          name = "kurtosis-cloud-backend";
          tag = rev;
          created = "now";
          contents = server;
          architecture = "arm64";
          config.Cmd = [ "${server}/bin/linux_arm64/server" ];
        };

        testingVMs = import ./internal_testsuites/vm_modules.nix {
          inherit nixpkgs;
          engine_image = container.image.arm64;
        };

        packages.x86_64-linux.linuxVM =
          testingVMs.nixosConfigurations.linuxVM.config.system.build.vm;

        packages.aarch64-darwin.darwinVM =
          testingVMs.nixosConfigurations.darwinVM.config.system.build.vm;

        packages.testVM = import ./internal_testsuites/vm_tests.nix
          (self.inputs // { inherit pkgs nixpkgs; });
      });
}
