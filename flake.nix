{
  description = "Go dev env";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        devShells.default = pkgs.mkShell {
          hardeningDisable = ["fortify"];
          packages = with pkgs; [
            go_1_25
            gopls
            gotools
            golangci-lint
            delve
            git
            gnumake
            pkg-config
          ];

          shellHook = ''
            export GOPATH="$PWD/.gopath"
            export GOBIN="$GOPATH/bin"
            export PATH="$GOBIN:$PATH"
            export GOPROXY="https://goproxy.cn,direct"
            export GOSUMDB="sum.golang.google.cn"
            export GOPRIVATE=""
            export GONOSUMDB="$GOPRIVATE"
            export GONOPROXY="$GOPRIVATE"
            export GO111MODULE="on"
            mkdir -p "$GOPATH" "$GOBIN"

            echo "[devShell] go=$(go version)"
            echo "[devShell] GOPROXY=$GOPROXY"
          '';
        };
      }
    );
}
