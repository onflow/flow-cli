#!/bin/sh

# Exit as soon as any command fails
set -e

REPO="onflow/flow-cli"
ASSETS_URL="https://github.com/$REPO/releases/download/"
# The version to download, set by get_version (defaults to args[1])
VERSION="$1"
# The architecture string, set by get_architecture
ARCH=""
# The tag search term to use if no version is specified (first match is used)
SEARCH_TERM="cadence-v1.0.0"

# Optional environment variable for Github API token
# If GITHUB_TOKEN is set, use it in the curl requests to avoid rate limiting
github_token_header=""
if [ -n "$GITHUB_TOKEN" ]; then
  github_token_header="Authorization: Bearer $GITHUB_TOKEN"
fi

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
            _cputype=amd64
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
page=1
get_version() {
  if [ -z "$VERSION" ]
  then
    VERSION=""
    if [ -n "$github_token_header" ]
    then
      response=$(curl -H "$github_token_header" -s "https://api.github.com/repos/$REPO/releases?per_page=100&page=$page" -w "%{http_code}")
    else
      response=$(curl -s "https://api.github.com/repos/$REPO/releases?per_page=100&page=$page" -w "%{http_code}")
    fi

    status=$(echo "$response" | tail -n 1)
    if [ "$status" -eq "403" ] && [ -n "$github_token_header" ]
    then
      echo "Failed to get latest version from Github API, is your GITHUB_TOKEN valid?  Trying without authentication..."
      github_token_header=""
      get_version
    fi

    if [ "$status" -ne "200" ]
    then
      echo "Failed to get latest version from Github API, please manually specify a version to install as an argument to this script."
      return 1
    fi

    VERSION=$(echo "$response" | grep -E 'tag_name' | grep -E "$SEARCH_TERM" | head -n 1 | cut -d '"' -f 4)

    if [ -z "$VERSION" ]
    then
      page=$((page+1))
      get_version
    fi
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
  if [ -n "$github_token_header" ]
  then
    curl -H "$github_token_header" -L --progress-bar "$url" -o $tmpfile
  else
    curl -L --progress-bar "$url" -o $tmpfile
  fi

  # Ensure we don't receive a not found error as response.
  if grep -q "Not Found" $tmpfile
  then
    echo "Version $VERSION could not be found"
    exit 1
  fi

  [ -d $TARGET_PATH ] || mkdir -p $TARGET_PATH

  tar -xf $tmpfile -C $TARGET_PATH
  mv $TARGET_PATH/flow-cli $TARGET_PATH/flow-c1
  chmod +x $TARGET_PATH/flow-c1

  echo ""
  echo "Successfully installed Flow CLI $VERSION to $TARGET_PATH."
  echo "Make sure $TARGET_PATH is in your \$PATH environment variable."
  echo ""
  echo "PRE-RELEASE: Use the 'flow-c1' command to interact with this Cadence 1.0 CLI pre-release."
  echo ""
}

main
