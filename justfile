pi_host := "homie"
pi_dir := "/opt/homie"

# Run the API (with hot reload) and both Vite dev servers in Docker.
dev:
    docker compose up --build

# Cross-compile the homie-api binary (linux/arm64) with both SPAs embedded.
build:
    docker buildx build --platform linux/arm64 --target export --output type=local,dest=./dist .

# Build and ship the binary + kiosk URL to the Pi, then restart services.
deploy: build
    rsync -az ./dist/homie-api {{pi_host}}:{{pi_dir}}/homie-api.new
    ssh {{pi_host}} "mv {{pi_dir}}/homie-api.new {{pi_dir}}/homie-api && chmod +x {{pi_dir}}/homie-api"
    ssh {{pi_host}} "sudo systemctl restart homie-api.service"

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
