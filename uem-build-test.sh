#!/bin/sh
################################################################################
# Copyright (c) 2024-2026 Tenebris Technologies Inc.                           #
# Please see the LICENSE file for details                                      #
################################################################################

#
# If moving the script outside the repo, update REPO below
REPO="$(cd "$(dirname "$0")" && pwd)"
#
# Default location for server and cli binaries
BIN_DIR="/tmp/uem-test-build"
#
# Default location of http directory for client distribution 
HTTP_DIR="/tmp/uem-test-build"
#
# Folder to build into ($REPO/bin is excluded in .gitignore)
BUILD_DIR="/tmp/uem-test-build"
#
# Build options
BOPTS="-ldflags=\"-s -w\""
#
# Required minimum version
GO_MIN_VERSION="1.25"
#
# Bail if an error occurs
set -e
#
rm -rf /tmp/uem-test-build
mkdir /tmp/uem-test-build
#
# Make sure govulncheck is up to date
go install golang.org/x/vuln/cmd/govulncheck@latest
#
################################################################
# uem-agent has conditional code for Windows, Linux, and macOS.
# Running vet and govulncheck for each build is slower, but it
# makes sure that issues in conditional code are detected.
###############################################################
#
build() {
  os=$1
  arch=$2
  out=$3

  echo ""
  echo "Building for $os $arch..."
  echo "go vet..."
  GOOS=$os GOARCH=$arch go vet ./...
  if [ $? -ne 0 ]; then
    echo ""
    echo "*** go vet failed - aborting ***"
    echo ""
    exit 1
  fi

  echo "govulncheck..."
  GOOS=$os GOARCH=$arch govulncheck ./...
  if [ $? -ne 0 ]; then
    echo ""
    echo "*** govulncheck failed - aborting ***"
    echo ""
    exit 1
  fi

  BIN="$out-$os-$arch"
  if [ "$os" = "windows" ]; then
    BIN="${BIN}.exe"
  fi
  echo "Compiling for $os $arch to $BUILD_DIR..."
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build $BOPS -o $BUILD_DIR/$BIN
}

#
# Check if a directory exists 
# with /opt/uem-server
# 
check_directory() {
    if  ! test -d "$1"; then
	echo ""
        echo "Error: Directory does not exist: $1" >&2
	echo "Please correct this issue and try again."
	echo ""
        exit 1
    fi
}

check_go_version() {
    # Get the Go version (e.g., "go version go1.24.0 linux/amd64")
    GO_VERSION_OUTPUT=$(go version)

    # Extract the version number (e.g., "1.24.0")
    GO_VERSION=$(echo "$GO_VERSION_OUTPUT" | awk '{print $3}' | sed 's/go//')

    # Compare versions
    if [ "$(printf '%s\n' "$GO_VERSION" "$REQUIRED_VERSION" | sort -V | head -n 1)" != "$REQUIRED_VERSION" ]; then
        echo "Error: Current Go version is $GO_VERSION."
        echo "UnifyEM is designed for Go $REQUIRED_VERSION or later. Please upgrade Go and try again." >&2
        exit 1
    fi
    echo "Go version $GO_VERSION is installed."
}

#
################################################################
# Make sure go is installed and key directories exist
################################################################
#
if ! command -v go >/dev/null 2>&1; then
    echo "Error: go is not installed. Please install Go and try again." >&2
    exit 1
fi
check_go_version

if ! command -v govulncheck >/dev/null 2>&1; then
    echo "Error: govulncheck is not installed." >&2
    echo "Attempting to install 'golang.org/x/vuln/cmd/govulncheck@latest'..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
    exit 1
fi

check_directory $REPO
check_directory $BIN_DIR
check_directory $HTTP_DIR
#
################################################################
# Build uem-server and uem-cli and copy to $BIN_DIR
################################################################
#
echo "Changing to $REPO..."
cd $REPO
mkdir -p $BUILD_DIR
echo ""
echo "---"
echo "Building uem-server to $BUILD_DIR..."
echo ""
cd $REPO/server
build windows amd64 uem-server
build windows arm64 uem-server
build windows 386 uem-server
build darwin amd64 uem-server
build darwin arm64 uem-server
build linux amd64 uem-server
build linux arm64 uem-server
build linux 386 uem-server

echo ""
echo "---"
echo "Building uem-cli to $BUILD_DIR..."
echo ""
cd $REPO/cli
build windows amd64 uem-cli
build windows arm64 uem-cli
build windows 386 uem-cli
build darwin amd64 uem-cli
build darwin arm64 uem-cli
build linux amd64 uem-cli
build linux arm64 uem-cli
build linux 386 uem-cli
#
################################################################
# Build agents and copy to $BIN_DIR
################################################################
#
echo ""
echo "---"
echo "Building agents to $BUILD_DIR..."
echo ""
cd $REPO/agent
build windows 386 uem-agent
build windows amd64 uem-agent
build windows arm64 uem-agent
build darwin amd64 uem-agent
build darwin arm64 uem-agent
build linux 386 uem-agent
build linux amd64 uem-agent
build linux arm64 uem-agent
echo ""
echo "Finished building agents"
echo ""
echo "Test build complete"
