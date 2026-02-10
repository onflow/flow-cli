{
  description = "Flow CLI - Command-line interface for the Flow blockchain";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        # Version detection:
        # - When building from a git tag (e.g., nix build github:onflow/flow-cli/v2.14.2),
        #   the version is automatically extracted from the tag: "2.14.2"
        # - For local development builds, version is "dev"
        # This eliminates the need for a separate VERSION file and keeps version
        # management in sync with git tags used for releases.
        version =
          if (self ? ref) && (builtins.match "v[0-9]+\..+" self.ref != null)
          then builtins.substring 1 (-1) self.ref  # Remove 'v' prefix from version tags
          else "dev";

        shortRev = self.shortRev or "dev";
      in
      {
        packages = {
          flow-cli = pkgs.buildGoModule {
            pname = "flow-cli";
            version = version;
            src = ./.;

            vendorHash = "sha256-EYQfXvHiRftod45Rvi7dUHF+3G5PyDtdM+HmJsE5r4I=";
            proxyVendor = true;

            subPackages = [ "cmd/flow" ];

            env = {
              CGO_ENABLED = "1";
              CGO_CFLAGS = "-O2 -D__BLST_PORTABLE__";
            };

            ldflags = [
              "-s" "-w"
              "-X github.com/onflow/flow-cli/build.semver=v${version}"
              "-X github.com/onflow/flow-cli/build.commit=${shortRev}"
              "-X github.com/onflow/flow-cli/internal/accounts.accountToken=lilico:sF60s3wughJBmNh2"
              "-X github.com/onflow/flow-cli/internal/command.MixpanelToken=3fae49de272be1ceb8cf34119f747073"
            ];

            preCheck = ''
              export SKIP_NETWORK_TESTS=1
            '';

            meta = with pkgs.lib; {
              description = "Command-line interface for the Flow blockchain";
              homepage = "https://developers.flow.com/tools/flow-cli";
              license = licenses.asl20;
              mainProgram = "flow";
            };
          };

          default = self.packages.${system}.flow-cli;
        };

        apps = {
          flow-cli = flake-utils.lib.mkApp {
            drv = self.packages.${system}.flow-cli;
            name = "flow";
          };
          default = self.apps.${system}.flow-cli;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            gotools
            gopls
            delve
            gnumake
            git
          ];

          CGO_ENABLED = 1;
          CGO_CFLAGS = "-O2 -D__BLST_PORTABLE__";
        };
      }
    );
}
