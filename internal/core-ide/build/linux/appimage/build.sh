#!/usr/bin/env bash
# Copyright (c) 2018-Present Lea Anthony
# SPDX-License-Identifier: MIT

# Fail script on any error
set -euxo pipefail

# Define variables
APP_DIR="${APP_NAME}.AppDir"

# Create AppDir structure
mkdir -p "${APP_DIR}/usr/bin"
cp -r "${APP_BINARY}" "${APP_DIR}/usr/bin/"
cp "${ICON_PATH}" "${APP_DIR}/"
cp "${DESKTOP_FILE}" "${APP_DIR}/"

ARCH=$(uname -m)
case "${ARCH}" in
    x86_64)
        DEPLOY_ARCH="x86_64"
        ;;
    aarch64|arm64)
        DEPLOY_ARCH="aarch64"
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}" >&2
        exit 1
        ;;
esac

# Download linuxdeploy and make it executable
wget -q -4 -N "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-${DEPLOY_ARCH}.AppImage"
chmod +x "linuxdeploy-${DEPLOY_ARCH}.AppImage"

# Run linuxdeploy to bundle the application
"./linuxdeploy-${DEPLOY_ARCH}.AppImage" --appdir "${APP_DIR}" --output appimage

# Rename the generated AppImage (glob must be unquoted)
mv ${APP_NAME}*.AppImage "${APP_NAME}.AppImage"

