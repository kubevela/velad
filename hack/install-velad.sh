#!/usr/bin/env bash

# Implemented based on Dapr Cli https://github.com/dapr/cli/tree/master/install

# VelaD location
: ${VELAD_INSTALL_DIR:="/usr/local/bin"}

# sudo is required to copy binary to VELAD_INSTALL_DIR for linux
: ${USE_SUDO:="false"}

# Http request CLI
VELAD_HTTP_REQUEST_CLI=curl

# VelaD filename
VELAD_CLI_FILENAME=velad

VELAD_CLI_FILE="${VELAD_INSTALL_DIR}/${VELAD_CLI_FILENAME}"

DOWNLOAD_BASE="https://static.kubevela.net/binary/velad"

getSystemInfo() {
    ARCH=$(uname -m)
    case $ARCH in
        armv7*) ARCH="arm";;
        aarch64) ARCH="arm64";;
        x86_64) ARCH="amd64";;
    esac

    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    # Most linux distro needs root permission to copy the file to /usr/local/bin
    if [ "$OS" == "linux" ] || [ "$OS" == "darwin" ]; then
        if [ "$VELAD_INSTALL_DIR" == "/usr/local/bin" ]; then
            USE_SUDO="true"
        fi
    fi
}

verifySupported() {
    local supported=(darwin-amd64 darwin-arm64 linux-amd64 linux-arm linux-arm64)
    local current_osarch="${OS}-${ARCH}"

    for osarch in "${supported[@]}"; do
        if [ "$osarch" == "$current_osarch" ]; then
            echo "Your system is ${OS}_${ARCH}"
            return
        fi
    done

    echo "No prebuilt binary for ${current_osarch}"
    exit 1
}

runAsRoot() {
    local CMD="$*"

    if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
        CMD="sudo $CMD"
    fi

    $CMD
}

checkHttpRequestCLI() {
    if type "curl" > /dev/null; then
        VELAD_HTTP_REQUEST_CLI=curl
    elif type "wget" > /dev/null; then
        VELAD_HTTP_REQUEST_CLI=wget
    else
        echo "Either curl or wget is required"
        exit 1
    fi
}

checkExistingVelaD() {
    if [ -f "$VELAD_CLI_FILE" ]; then
        echo -e "\nVelaD is detected:"
        $VELAD_CLI_FILE version
        echo -e "Reinstalling VelaD - ${VELAD_CLI_FILE}...\n"
    else
        echo -e "Installing VelaD ...\n"
    fi
}

getLatestRelease() {
    local velaReleaseUrl="${DOWNLOAD_BASE}/latest_version"
    local latest_release=""

    if [ "$VELAD_HTTP_REQUEST_CLI" == "curl" ]; then
        latest_release=$(curl -s $velaReleaseUrl)
    else
        latest_release=$(wget -q -O - $velaReleaseUrl)
    fi

    ret_val=$latest_release
}

downloadFile() {
    LATEST_RELEASE_TAG=$1

    VELA_CLI_ARTIFACT="${VELAD_CLI_FILENAME}-${OS}-${ARCH}-${LATEST_RELEASE_TAG}.tar.gz"
    # convert `-` to `_` to let it work
    DOWNLOAD_URL="${DOWNLOAD_BASE}/${LATEST_RELEASE_TAG}/${VELA_CLI_ARTIFACT}"

    # Create the temp directory
    VELAD_TMP_ROOT=$(mktemp -dt velad-install-XXXXXX)
    ARTIFACT_TMP_FILE="$VELAD_TMP_ROOT/$VELA_CLI_ARTIFACT"

    echo "Downloading $DOWNLOAD_URL ..."
    if [ "$VELAD_HTTP_REQUEST_CLI" == "curl" ]; then
        curl -SsL "$DOWNLOAD_URL" -o "$ARTIFACT_TMP_FILE"
    else
        wget -q -O "$ARTIFACT_TMP_FILE" "$DOWNLOAD_URL"
    fi

    if [ ! -f "$ARTIFACT_TMP_FILE" ]; then
        echo "failed to download $DOWNLOAD_URL ..."
        exit 1
    fi
}

installFile() {
    tar xf "$ARTIFACT_TMP_FILE" -C "$VELAD_TMP_ROOT"
    local tmp_root_velad="$VELAD_TMP_ROOT/${OS}-${ARCH}/$VELAD_CLI_FILENAME"

    if [ ! -f "$tmp_root_velad" ]; then
        echo "Failed to unpack VelaD executable."
        exit 1
    fi

    chmod o+x "$tmp_root_velad"
    runAsRoot cp "$tmp_root_velad" "$VELAD_INSTALL_DIR"

    if [ $? -eq 0 ] && [ -f "$VELAD_CLI_FILE" ]; then
        echo "VelaD installed into $VELAD_INSTALL_DIR/$VELAD_CLI_FILENAME successfully."
        echo ""
        $VELAD_CLI_FILE version
    else
        echo "Failed to install $VELAD_CLI_FILENAME"
        exit 1
    fi
}

fail_trap() {
    result=$?
    if [ "$result" != "0" ]; then
        echo "Failed to install VelaD"
        echo "Go to https://kubevela.io for more support."
    fi
    cleanup
    exit $result
}

cleanup() {
    if [[ -d "${VELAD_TMP_ROOT:-}" ]]; then
        rm -rf "$VELAD_TMP_ROOT"
    fi
}

installCompleted() {
    echo -e "\nFor more information on how to started, please visit:"
    echo -e "  https://kubevela.io"
}

# -----------------------------------------------------------------------------
# main
# -----------------------------------------------------------------------------
trap "fail_trap" EXIT

getSystemInfo
verifySupported
checkExistingVelaD
checkHttpRequestCLI


if [ -z "$1" ]; then
    echo "Getting the latest VelaD..."
    getLatestRelease
elif [[ $1 == v* ]]; then
    ret_val=$1
else
    ret_val=v$1
fi

downloadFile $ret_val
installFile
cleanup

installCompleted