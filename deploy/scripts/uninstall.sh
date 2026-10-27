#!/bin/sh
# Fully removes Homie (services, config, data) from this Pi.
# Usage: sudo sh /usr/share/homie/uninstall.sh [--yes]
#
# The body lives in main() and is only called on the last line: purging the
# package deletes this file, and the shell must have parsed all of it first.
set -eu

main() {
    assume_yes=0
    for arg in "$@"; do
        case "$arg" in
            -y|--yes) assume_yes=1 ;;
            *)
                echo "usage: sudo sh $0 [--yes]" >&2
                exit 2
                ;;
        esac
    done

    if [ "$(id -u)" -ne 0 ]; then
        echo "homie-uninstall: must run as root: sudo sh $0" >&2
        exit 1
    fi

    if [ "$assume_yes" -ne 1 ]; then
        echo "This permanently removes Homie, including its config in /etc/homie"
        echo "and its data in /var/lib/homie (database, WeatherKit key)."
        printf "Continue? [y/N] "
        read -r answer || answer=""
        case "$answer" in
            y|Y|yes|YES) ;;
            *)
                echo "Aborted."
                exit 1
                ;;
        esac
    fi

    systemctl stop homie-kiosk.service homie-api.socket homie-api.service homie-emoji-font.service homie-emoji-font.timer 2>/dev/null || true
    systemctl disable homie-kiosk.service homie-api.socket homie-api.service homie-emoji-font.service homie-emoji-font.timer 2>/dev/null || true
    systemctl start getty@tty1.service 2>/dev/null || true

    if dpkg -s homie >/dev/null 2>&1; then
        dpkg --purge homie
    else
        rm -f /usr/bin/homie-api \
            /lib/systemd/system/homie-api.service \
            /lib/systemd/system/homie-api.socket \
            /lib/systemd/system/homie-kiosk.service \
            /lib/systemd/system/homie-emoji-font.service \
            /lib/systemd/system/homie-emoji-font.timer
    fi

    # Fetched at runtime by homie-emoji-font.service, so not dpkg's to remove.
    rm -f /usr/local/share/fonts/AppleColorEmoji.ttf
    fc-cache --force /usr/local/share/fonts >/dev/null 2>&1 || true

    rm -rf /etc/homie \
        /var/lib/homie \
        /var/cache/homie \
        /usr/share/homie \
        /etc/systemd/system/homie-kiosk.service.d
    rm -f /etc/avahi/services/homie.service \
        /etc/sudoers.d/homie-deploy

    systemctl daemon-reload 2>/dev/null || true
    systemctl try-restart avahi-daemon.service 2>/dev/null || true

    echo "Homie has been removed."
    echo "Left in place on purpose:"
    echo "  - the 'homie' user (remove with: sudo userdel -r homie)"
    echo "  - the 'homie' hostname and its 127.0.1.1 line in /etc/hosts"
    echo "    (change with: sudo hostnamectl set-hostname NAME, then edit /etc/hosts)"
}

main "$@"
