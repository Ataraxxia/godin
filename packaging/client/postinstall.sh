#!/bin/sh
set -e

case "$1" in
    configure)  [ -z "$2" ] && action=install || action=upgrade ;;
    1)          action=install ;;
    2)          action=upgrade ;;
    *)          action=install ;;
esac

systemctl daemon-reload >/dev/null 2>&1 || :

if [ "$action" = "install" ]; then
    systemctl preset godin-server.service >/dev/null 2>&1 || :
    echo "Edit /etc/godin/godin-client.conf, then: systemctl enable --now godin-client"
else
    systemctl try-restart godin-server.service >/dev/null 2>&1 || :
fi
