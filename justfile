pi_host := "homie"
pi_dir := "/opt/homie"

# Run the API (with hot reload) and both Vite dev servers in Docker.
dev:
    docker compose up --build

# Cross-compile the homie-api binary (linux/arm64) with both SPAs embedded.
build:
    docker buildx build --platform linux/arm64 --target export --output type=local,dest=./dist .

# The SPAs are embedded in the binary, so the kiosk browser restarts too:
# without it Chromium keeps rendering the page it loaded before the deploy
# and frontend changes appear to have no effect. The two restarts stay
# separate calls: the Pi's NOPASSWD sudo rules match one exact unit each,
# so restarting both in a single systemctl call would prompt for a password.

# Build and ship the binary to the Pi, then restart the API and the kiosk.
deploy: build
    rsync -az ./dist/homie-api {{pi_host}}:{{pi_dir}}/homie-api.new
    ssh {{pi_host}} "mv {{pi_dir}}/homie-api.new {{pi_dir}}/homie-api && chmod +x {{pi_dir}}/homie-api"
    ssh {{pi_host}} "sudo systemctl restart homie-api.service"
    ssh {{pi_host}} "sudo systemctl restart homie-kiosk.service"

# The theme must land under the name "default": cage asks wlroots for a NULL
# theme, which resolves to "default" and never consults XCURSOR_THEME. The
# cursors live in {{pi_dir}}, owned by the homie user, so this needs no
# password. Only needed once per device, or after reimaging.

# Install the transparent cursor theme that the kiosk unit's XCURSOR_PATH points at.
install-blank-cursor:
    rsync -az ./deploy/blank-cursor.py {{pi_host}}:{{pi_dir}}/blank-cursor.py
    ssh {{pi_host}} "python3 {{pi_dir}}/blank-cursor.py {{pi_dir}}/cursors default"
    ssh {{pi_host}} "sudo systemctl restart homie-kiosk.service"

# Push updated systemd units (prompts for a password).
deploy-units:
    rsync -az ./deploy/*.service {{pi_host}}:/tmp/
    ssh -t {{pi_host}} "sudo mv /tmp/homie-*.service /etc/systemd/system/ && sudo systemctl daemon-reload"

# Point the kiosk browser at this Mac's Vite dev server.
kiosk-dev:
    #!/usr/bin/env bash
    set -euo pipefail
    mac_host="$(scutil --get LocalHostName)"
    ssh {{pi_host}} "echo HOMIE_KIOSK_URL=http://${mac_host}.local:5173 > {{pi_dir}}/kiosk.env"
    ssh {{pi_host}} "sudo systemctl restart homie-kiosk.service"

# Point the kiosk browser back at the on-device API.
kiosk-prod:
    ssh {{pi_host}} "echo HOMIE_KIOSK_URL=http://localhost:3000 > {{pi_dir}}/kiosk.env"
    ssh {{pi_host}} "sudo systemctl restart homie-kiosk.service"
