#!/bin/sh

# Exit as soon as any command fails
set -e

REPO="onflow/flow-cli"
ASSETS_URL="https://github.com/$REPO/releases/download/"
# The version to install (defaults to args[1])
VERSION="$1"
# The Cadence 1.0 version to install (defaults to args[2])
C1VERSION="$2"
# The architecture string, set by get_architecture
ARCH=""

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

get_version() {
    local search_term="$1"
    local page="$2"

    local version=""

    response=$(curl -H "$github_token_header" -s "https://api.github.com/repos/$REPO/releases?per_page=10&page=$page" -w "%{http_code}")

    status=$(echo "$response" | tail -n 1)
    if [ "$status" -eq "403" ] && [ -n "$github_token_header" ]
    then
      echo "Failed to get releases from Github API, is your GITHUB_TOKEN valid? Re-trying without authentication ..."
      github_token_header=""
      get_version "$search_term" "$page"
    fi

    if [ "$status" -ne "200" ]
    then
      echo "Failed to get releases from Github API, please manually specify a version to install as an argument to this script."
      return 1
    fi

    version=$(echo "$response" | grep -E 'tag_name' | grep -E "$search_term" | head -n 1 | cut -d '"' -f 4)

    if [ -z "$version" ]
    then
      get_version "$search_term" "$((page+1))"
    fi

    echo "$version"
}

get_latest() {
    local version=""

    response=$(curl -H "$github_token_header" -s "https://api.github.com/repos/$REPO/releases/latest" -w "%{http_code}")

    status=$(echo "$response" | tail -n 1)
    if [ "$status" -eq "403" ] && [ -n "$github_token_header" ]
    then
      echo "Failed to get latest release from Github API, is your GITHUB_TOKEN valid? Re-trying without authentication ..."
      github_token_header=""
      get_latest
    fi

    if [ "$status" -ne "200" ]
    then
      echo "Failed to get latest release from Github API, please manually specify a version to install as an argument to this script."
      return 1
    fi

    echo "$response" | grep -E 'tag_name' | grep -E "$search_term" | head -n 1 | cut -d '"' -f 4
}

# Function to download and install a specified version
install_version() {
  local version="$1"
  local target_name="$2"

  echo "Installing version $version ..."

  tmpfile=$(mktemp 2>/dev/null || mktemp -t flow)
  url="$ASSETS_URL$version/flow-cli-$version-$ARCH.tar.gz"
  curl -H "$github_token_header" -L --progress-bar "$url" -o "$tmpfile"

  # Ensure we don't receive a not found error as response.
  if grep -q "Not Found" "$tmpfile"
  then
    echo "Version $version could not be found"
    exit 1
  fi

  [ -d "$TARGET_PATH" ] || mkdir -p "$TARGET_PATH"

  tar -xf "$tmpfile" -C "$TARGET_PATH"
  mv "$TARGET_PATH/flow-cli" "$TARGET_PATH/$target_name"
  chmod +x "$TARGET_PATH/$target_name"
}

# Determine the system architecture, download the appropriate binaries, and
# install them in `/usr/local/bin` on macOS and `~/.local/bin` on Linux
# with executable permissions.
main() {
  get_architecture || exit 1

  if [ -z "$VERSION" ]
  then
    echo "Getting version of latest stable release ..."

    VERSION=$(get_latest || exit 1)
  fi

  install_version "$VERSION" "flow"

  if [ -z "$C1VERSION" ]
  then
    echo "Getting version of latest Cadence 1.0 preview release ..."

    C1VERSION=$(get_version "cadence-v1.0.0" 1 || exit 1)
  fi

  install_version "$C1VERSION" "flow-c1"

  echo ""
  echo "Successfully installed Flow CLI $VERSION as 'flow'."
  echo "Use the 'flow' command to interact with the Flow CLI compatible with versions of Cadence before 1.0 (only)."
  echo ""
  echo "Successfully installed Flow CLI $C1VERSION as 'flow-c1'."
  echo "Use the 'flow-c1' command to interact with the Flow CLI preview compatible with Cadence 1.0 (only)."
  echo ""

}

main