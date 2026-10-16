#!/usr/bin/env bash
# One-time Pi bootstrap. Idempotent: safe to re-run.
# Requires a password the first time (installs packages, units, sudoers).
# Run on the Pi itself, e.g.:
#   ssh homie 'bash -s' < deploy/install.sh
set -euo pipefail

sudo apt-get update
sudo apt-get install -y cage fonts-noto-color-emoji

sudo install -d -o homie -g homie -m 0755 /opt/homie

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

sudo cp "$SCRIPT_DIR/homie-api.service" /etc/systemd/system/homie-api.service
sudo cp "$SCRIPT_DIR/homie-kiosk.service" /etc/systemd/system/homie-kiosk.service

sudo visudo -cf "$SCRIPT_DIR/sudoers.d/homie-deploy"
sudo cp "$SCRIPT_DIR/sudoers.d/homie-deploy" /etc/sudoers.d/homie-deploy
sudo chmod 0440 /etc/sudoers.d/homie-deploy

sudo systemctl daemon-reload
sudo systemctl enable homie-api.service homie-kiosk.service

echo "Bootstrap complete. Deploy a binary with 'task deploy', then:"
echo "  sudo systemctl start homie-api.service homie-kiosk.service"
