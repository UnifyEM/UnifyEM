#!/bin/sh
#
# If moving the script outside the repo, update REPO below
REPO="$(cd "$(dirname "$0")" && pwd)"
#
# Default location for server and cli binaries
BIN_DIR="/usr/local/bin"
#
# Default location of http directory for client distribution 
HTTP_DIR="/opt/uem-server/http"
#
# Folder to build into ($REPO/bin is excluded in .gitignore)
BUILD_DIR="$REPO/bin"
#
# Build options
BOPTS="-ldflags=\"-s -w\""
#
# Required minimum version
GO_MIN_VERSION="1.24"
#
# Bail if an error occurs
set -e
#
################################################################
# uem-agent has conditional code for Windows, Linux, and macOS.
# Running vet and govulncheck for each build is slower, but it
# makes sure that issues in conditional code are detected.
###############################################################
#
build_agent() {
  os=$1
  arch=$2

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

  BIN="uem-agent-$os-$arch"
  if [ "$os" = "windows" ]; then
    BIN="${BIN}.exe"
  fi
  echo "Compiling uem-agent for $os $arch to $BUILD_DIR..."
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build $BOPS -o $BUILD_DIR/$BIN
}

#
# Check if a directory exists - use sudo to prevent issues 
# with /opt/uem-server
# 
check_directory() {
    if  ! sudo test -d "$1"; then
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
echo "Executing git pull..."
git pull
echo "Building uem-server to $BUILD_DIR..."
cd $REPO/server
CGO_ENABLED=0 go build -o $BUILD_DIR/uem-server
echo "Building uem-cli to $BUILD_DIR..."
cd $REPO/cli
CGO_ENABLED=0 go build -o $BUILD_DIR/uem-cli
echo "Copying cli binary to $BIN_DIR..."
sudo cp $BUILD_DIR/uem-cli $BIN_DIR/uem-cli
echo "Stopping uem-server..."
sudo systemctl stop uem-server
echo "Copying server binary to $BIN_DIR..."
sudo cp $BUILD_DIR/uem-server $BIN_DIR/uem-server
echo "Starting server..."
sudo systemctl daemon-reload
sudo systemctl start uem-server
#sleep 1
#echo ""
#echo "Checking service status..."
#sudo systemctl status uem-server
#
################################################################
# Build uem-server and uem-cli and copy to $BIN_DIR
################################################################
#
echo ""
echo "Building agents to $BUILD_DIR..."
cd $REPO/agent
build_agent windows 386
build_agent windows amd64
build_agent windows arm64
build_agent darwin amd64
build_agent darwin arm64
build_agent linux 386
build_agent linux amd64
build_agent linux arm64
echo ""
echo "Finished building agents"
echo "Copying agents $HTTP_DIR..."
sudo cp -f $BUILD_DIR/uem-agent-* $HTTP_DIR
echo ""
sudo ls -al $HTTP_DIR
echo ""
echo "Attempting to create deployment file..."
$BIN_DIR/uem-cli files deploy
echo ""
echo ""
echo "Agent upgrade notes:"
echo "Use 'uem-cli agent list' to obtain a list of agents"
echo ""
echo "Use 'uem-cli cmd upgrade agent_id=<agent ID>' to initiate an agent upgrade"
echo ""
echo "For legacy agents, add 'hash=false' to the upgrade command to suppress the hash because it will cause"
echo "older agents to reject the command."
echo ""
echo "Caution: 'uem-cli files deploy' must run to update the hashes in the deployment file or agents will not upgrade."
echo "          Verify that it ran correctly above, otherwise run it manually."
echo ""
