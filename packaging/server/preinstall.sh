#!/bin/sh
set -e

if command -v systemd-sysusers >/dev/null 2>&1 && \
   [ -f /usr/lib/sysusers.d/godin.conf ]; then
    systemd-sysusers /usr/lib/sysusers.d/godin.conf >/dev/null 2>&1 || :
fi

if ! getent group godin >/dev/null; then
    groupadd --system godin
fi
if ! getent passwd godin >/dev/null; then
    useradd --system --gid godin --home-dir /var/lib/godin \
            --no-create-home --shell /usr/sbin/nologin \
            --comment "GOdin monitoring server" godin
fi
