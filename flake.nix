{
  description = "0KN";

  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixos-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs = { nixpkgs, flake-utils, ... }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
      };

      libmcl = (with pkgs; stdenv.mkDerivation {
        pname = "libmcl";
        version = "1.86.0";
        src = fetchgit {
          url = "https://github.com/herumi/mcl";
          rev = "v1.86.0";
          sha256 = "wh7VM1AYcDkNSKKZ4R0RirwlIT83NpLMMnhpdt2R78E=";
        };
        nativeBuildInputs = [
          clang
          cmake
        ];
        buildInputs = [
          gmp
        ];
        buildPhase = "make -j $NIX_BUILD_CORES";
        installPhase = "PREFIX=$out make install";
      });

    in rec {
      defaultApp = flake-utils.lib.mkApp {
        drv = defaultPackage;
      };
      defaultPackage = libmcl;
      devShell = pkgs.mkShell {
        buildInputs = with pkgs; [
          libmcl
          go
        ];
      };
    }
  );
}
