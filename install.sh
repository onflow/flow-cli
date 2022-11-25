#!/bin/sh

# Exit as soon as any command fails
set -e

REPO="sideninja/flow-cli"
GITHUB_URL="https://api.github.com/repos/$REPO"
BASE_URL="https://storage.googleapis.com/flow-cli"
ASSETS_URL="https://github.com/$REPO/releases/download/"
# The version to download, set by get_version (defaults to args[1])
VERSION="$1"
# The architecture string, set by get_architecture
ARCH=""

# Get the architecture (CPU, OS) of the current system as a string.
# Only MacOS/x86_64/ARM64 and Linux/x86_64/ARM64 architectures are supported.
get_architecture() {
    _ostype="$(uname -s)"
    _cputype="$(uname -m)"
    _targetpath=""
    if [ "$_ostype" = Darwin ] && [ "$_cputype" = i386 ]; then
        if sysctl hw.optional.x86_64 | grep -q ': 1'; then
            _cputype=x86_64
        fi
    fi
    case "$_ostype" in
        Linux)
            _ostype=linux
            _targetpath=$HOME/.local/bin
            ;;
        Darwin)
            _ostype=darwin
            _targetpath=/usr/local/bin
            ;;
        *)
            echo "unrecognized OS type: $_ostype"
            return 1
            ;;
    esac
    case "$_cputype" in
        x86_64 | x86-64 | x64 | amd64)
            _cputype=x86_64
            ;;
         arm64 | aarch64)
            _cputype=arm64
            ;;
        *)
            echo "unknown CPU type: $_cputype"
            return 1
            ;;
    esac
    _arch="${_ostype}-${_cputype}"
    ARCH="${_arch}"
    TARGET_PATH="${_targetpath}"
}

# Get the latest version from remote if none specified in args.
get_version() {
  if [ -z "$VERSION" ]
  then
    VERSION=$(curl -s "$GITHUB_URL/releases/latest" | grep -E 'tag_name' | cut -d '"' -f 4)
  fi
}

# Determine the system architecure, download the appropriate binary, and
# install it in `/usr/local/bin` on macOS and `~/.local/bin` on Linux
# with executable permission.
main() {

  get_architecture || exit 1
  get_version || exit 1

  echo "Downloading version $VERSION ..."

  tmpfile=$(mktemp 2>/dev/null || mktemp -t flow)
  url="$ASSETS_URL$VERSION/flow-cli-$VERSION-$ARCH.tar.gz"
  curl -L --progress-bar "$url" -o $tmpfile

  # Ensure we don't receive a not found error as response.
  if grep -q "Not Found" $tmpfile
  then
    echo "Version $VERSION could not be found"
    exit 1
  fi

  [ -d $TARGET_PATH ] || mkdir -p $TARGET_PATH

  tar -xf $tmpfile -C $TARGET_PATH
  mv $TARGET_PATH/flow-cli $TARGET_PATH/flow
  chmod +x $TARGET_PATH/flow

  echo "Successfully installed the Flow CLI to $TARGET_PATH."
  echo "Make sure $TARGET_PATH is in your \$PATH environment variable."
}

main
