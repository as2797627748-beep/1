#!/usr/bin/env bash
set -euo pipefail

APP_NAME=${APP_NAME:-autocode-platform}
REMOTE_HOST=${REMOTE_HOST:-}
REMOTE_PORT=${REMOTE_PORT:-22}
REMOTE_USER=${REMOTE_USER:-root}
REMOTE_BASE_DIR=${REMOTE_BASE_DIR:-/opt/autocode-platform}
ARCHIVE_NAME=${ARCHIVE_NAME:-release.tar.gz}
SSH_OPTS="-p ${REMOTE_PORT} -o ServerAliveInterval=30 -o ServerAliveCountMax=120"

if [[ -z "${REMOTE_HOST}" ]]; then
  printf 'REMOTE_HOST is required\n' >&2
  exit 1
fi

TIMESTAMP=$(date +%Y%m%d%H%M%S)
REMOTE_RELEASE_DIR="${REMOTE_BASE_DIR}/releases/${TIMESTAMP}"

# Build application archive locally before upload.
tar -czf "${ARCHIVE_NAME}" cmd internal scripts go.mod .monkeycode

# Prepare remote release directories without touching system files.
ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "mkdir -p '${REMOTE_BASE_DIR}/releases' '${REMOTE_BASE_DIR}/shared/logs'"

# Upload current release archive.
scp -P "${REMOTE_PORT}" "${ARCHIVE_NAME}" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_BASE_DIR}/${ARCHIVE_NAME}"

# Expand release in background-safe way and update current symlink.
ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "nohup bash -lc 'mkdir -p \"${REMOTE_RELEASE_DIR}\" && tar -xzf \"${REMOTE_BASE_DIR}/${ARCHIVE_NAME}\" -C \"${REMOTE_RELEASE_DIR}\" && ln -sfn \"${REMOTE_RELEASE_DIR}\" \"${REMOTE_BASE_DIR}/current\"' > '${REMOTE_BASE_DIR}/shared/logs/deploy.log' 2>&1 < /dev/null &"

printf 'Release uploaded. Follow remote log: %s\n' "${REMOTE_BASE_DIR}/shared/logs/deploy.log"
