#!/bin/sh

# Exit as soon as any command fails
set -e

REPO="onflow/flow-cli"
ASSETS_URL="https://github.com/$REPO/releases/download/"
# The version to download, set by get_version (defaults to args[1])
VERSION="$1"
# The architecture string, set by get_architecture
ARCH=""

# Optional environment variable for Github API token
# If GITHUB_TOKEN is set, use it in the curl requests to avoid rate limiting
github_token_header=""
if [ -n "$GITHUB_TOKEN" ]; then
  github_token_header="Authorization: Bearer $GITHUB_TOKEN"
fi

modify_dotfiles = true
if [ -n "$NO_MODIFY_DOTFILES" ]; then
  modify_dotfiles = false
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
            modify_dotfiles = false
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
get_version() {
  if [ -z "$VERSION" ]
  then
    VERSION=""
    if [ -n "$github_token_header" ]
    then
      VERSION=$(curl -H "$github_token_header" -s "https://api.github.com/repos/$REPO/releases/latest" | grep -E 'tag_name' | cut -d '"' -f 4)
    else
      VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -E 'tag_name' | cut -d '"' -f 4)
    fi

    if [ -z "$VERSION" ] && [ -n "$github_token_header" ]
    then
      echo "Failed to get latest version from Github API, is your GITHUB_TOKEN valid?  Trying without authentication..."
      github_token_header=""
      get_version
    fi
  fi
}


# Function to detect and append directory to the PATH variable
append_path_to_dotfiles() {
    # Do not modify the dotfiles if the user has set the NO_MODIFY_DOTFILES environment variable
    # or if dotfile modification was otherwise disabled by the script (i.e. on macOS)
    if [ "$modify_dotfiles" = false ]; then
        return
    fi

    # List of common profile files
    # Should support all POSIX compliant shells and Zsh
    local supported_dotfiles=("$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.zlogin")

    for dotfile in "${profile_files[@]}"; do
        if [[ -f "$dotfile" ]]; then
            # Check if the directory is already in the PATH
            if grep -q -x "export PATH=\$PATH:$1" "$profile_file"; then
                echo "Directory already in PATH in $profile_file."
            else
                # Append the directory to the PATH variable in the dotfile
                # It will only be appended if the directory is not already in the PATH
                echo ""                                         >> "$profile_file"
                echo "### FLOW CLI"                             >> "$profile_file"
                echo "case \":\$PATH:\" in"                     >> "$profile_file"
                echo " *\":$TARGET_PATH:\"*);;"                 >> "$profile_file"
                echo "  *)"                                     >> "$profile_file"
                echo "    export PATH=\"\$PATH:$TARGET_PATH\""  >> "$profile_file"
                echo "esac"                                     >> "$profile_file"
                echo "### END"                                  >> "$profile_file"
                ### END
                echo "Directory appended to PATH in $profile_file."
            fi
        fi
    done
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
  mv $TARGET_PATH/flow-cli $TARGET_PATH/flow
  chmod +x $TARGET_PATH/flow

  # Remove the temporary file
  rm $tmpfile

  # Add the directory to the PATH variable
  append_path_to_dotfiles "$TARGET_PATH"

  echo "Successfully installed the Flow CLI to $TARGET_PATH."
  echo "Make sure $TARGET_PATH is in your \$PATH environment variable."
}

main