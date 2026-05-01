#!/bin/sh
# Per-connection startup. ttyd execs this fresh for every WebSocket
# connection, so each visitor gets their own working copy of the
# template. The trap wipes the copy when the shell exits.

set -e

WORK=$(mktemp -d /tmp/shield-XXXXXX)
trap 'rm -rf "$WORK"' EXIT

cp -r /opt/shield-sandbox/template/. "$WORK/"
cd "$WORK"

clear
cat /etc/motd
echo
exec /bin/sh
