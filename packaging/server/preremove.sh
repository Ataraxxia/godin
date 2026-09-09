#!/bin/sh
set -e

case "$1" in
    remove|0)
        systemctl --no-reload disable --now godin-server.service >/dev/null 2>&1 || :
        ;;
esac
