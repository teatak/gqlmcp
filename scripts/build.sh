#!/usr/bin/env bash
set -e

# Ensure we are in the project root
cd "$(dirname "$0")/.." || exit 1

BIN="gqlmcp"
TYPE="release"

# Check for version file
if [ -f ".version" ]; then
    VERSION="$(cat .version)"
else
    VERSION="v0.0.0-dev"
    echo "Warning: .version file not found, using ${VERSION}"
fi

# Git info with fallback
if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    GITSHA="$(git rev-parse HEAD)"
    GITBRANCH="$(git rev-parse --abbrev-ref HEAD)"
else
    GITSHA="unknown"
    GITBRANCH="unknown"
fi

# Determine the arch/os combos we're building for
XC_OS=${XC_OS:-"linux darwin windows"}
XC_ARCH=${XC_ARCH:-"amd64 arm64"}

# -s: disable symbol table
# -w: disable DWARF generation
LDFLAGS="-s -w -X github.com/teatak/gqlmcp/common.Type=${TYPE} -X github.com/teatak/gqlmcp/common.GitSha=${GITSHA} -X github.com/teatak/gqlmcp/common.GitBranch=${GITBRANCH} -X github.com/teatak/gqlmcp/common.Version=${VERSION}"

# Delete the old dir
rm -rf bin/* pkg/*
mkdir -p bin/

if [ -n "${DEV}" ]; then
    XC_OS="$(go env GOOS)"
    XC_ARCH="$(go env GOARCH)"
fi

# Build!
echo "Start Building..."
FAILURES=0

for OS in ${XC_OS}; do
    for ARCH in ${XC_ARCH}; do
        # Simplified exclusion logic
        case "${OS}/${ARCH}" in
            darwin/arm|darwin/386|windows/arm|solaris/*)
                continue ;;
        esac

        echo "Building ${OS}/${ARCH}"
        NAME="${BIN}"
        if [ "${OS}" = "windows" ]; then
            NAME="${BIN}.exe"
        fi
        
        TARGET_DIR="./pkg/${OS}_${ARCH}"
        
        # Build with error reporting
        # We redirect stdout to /dev/null but keep stderr to see errors
        if ! CGO_ENABLED=0 GOOS="${OS}" GOARCH="${ARCH}" go build -ldflags "${LDFLAGS}" -o "${TARGET_DIR}/${NAME}" ./ > /dev/null; then
            echo -e "\033[31;1mBuilding ${OS}/${ARCH} error\033[0m"
            FAILURES=$((FAILURES+1))
        else
            # Local code signing for macOS to prevent "killed" error
            if [ "${OS}" = "darwin" ]; then
                if command -v codesign >/dev/null 2>&1; then
                    echo "Signing ${OS}/${ARCH} binary..."
                    codesign --force --sign - "${TARGET_DIR}/${NAME}"
                else
                    echo "Warning: codesign not found, skipping signing for ${OS}/${ARCH}"
                fi
            fi
        fi
    done
done

if [ "$FAILURES" -gt 0 ]; then
    echo -e "\033[31;1mBuild failed with $FAILURES errors.\033[0m"
    exit 1
fi

DEV_PLATFORM="./pkg/$(go env GOOS)_$(go env GOARCH)"
if [ -d "$DEV_PLATFORM" ]; then
    for F in $(find "${DEV_PLATFORM}" -mindepth 1 -maxdepth 1 -type f); do
        cp "${F}" ./bin/
    done
fi

# Zip and copy to the dist dir
for PLATFORM in ./pkg/*; do
    if [ -d "${PLATFORM}" ]; then
        OSARCH=$(basename "${PLATFORM}")
        echo "Packaging ${OSARCH}"

        pushd "${PLATFORM}" >/dev/null 2>&1
        zip -q -r "../${OSARCH}.zip" ./* >/dev/null 2>&1
        popd >/dev/null 2>&1
    fi
done

# Done!
echo -e "\033[32;1mBuild Done\033[0m"