pi_host := "homie.local"
pi_conf := "/etc/homie"
ios_simulator_id := env_var_or_default("IOS_SIMULATOR_ID", "6106C956-E47A-4099-81F5-851AB83AB540")
ios_derived_data := "/tmp/homie-ios-derived-data"
ios_app := ios_derived_data + "/Build/Products/Debug-iphonesimulator/Homie.app"

# Run the API (with hot reload) and both Vite dev servers in Docker.
dev:
    docker compose up --build

# Build the Photo Sync macOS test app.
photo-sync:
    ./apps/photo-sync/build.sh

# Build, install, and launch Homie in the iPhone 17 simulator.
# Override IOS_SIMULATOR_ID to target another booted or available simulator.
ios-sim:
    #!/usr/bin/env bash
    set -euo pipefail
    xcodebuild -project apps/mobile/ios/Homie.xcodeproj -scheme Homie -configuration Debug -sdk iphonesimulator -destination "platform=iOS Simulator,id={{ios_simulator_id}}" -derivedDataPath "{{ios_derived_data}}" CODE_SIGNING_ALLOWED=NO build
    xcrun simctl boot "{{ios_simulator_id}}" 2>/dev/null || true
    xcrun simctl bootstatus "{{ios_simulator_id}}" -b
    open -a Simulator --args -CurrentDeviceUDID "{{ios_simulator_id}}"
    xcrun simctl install "{{ios_simulator_id}}" "{{ios_app}}"
    xcrun simctl terminate "{{ios_simulator_id}}" com.schw.ar.Homie 2>/dev/null || true
    xcrun simctl launch "{{ios_simulator_id}}" com.schw.ar.Homie

# Remove local build artifacts (dist output and the iOS derived data).
clean:
    rm -rf dist apps/display/dist apps/manager/dist "{{ios_derived_data}}"

# Regenerate gqlgen code after editing services/api/graph/*.graphqls.
codegen:
    docker compose run --rm --no-deps api go run github.com/99designs/gqlgen generate

# Cross-compile the homie-api binary (linux/arm64) with both SPAs embedded.
build:
    docker buildx build --platform linux/arm64 --target export --output type=local,dest=./dist .

# Package the binary, systemd units, and blank-cursor theme as dist/homie.deb.
# The version is a UTC timestamp plus the commit, so a rebuild of the same
# commit still installs as an upgrade rather than being skipped as current.
package: build
    #!/usr/bin/env bash
    set -euo pipefail
    rm -rf dist/pkgroot dist/homie.deb
    mkdir -p dist/pkgroot/usr/bin dist/pkgroot/lib/systemd/system dist/pkgroot/etc/sudoers.d
    install -m 0755 dist/homie-api dist/pkgroot/usr/bin/homie-api
    install -m 0644 deploy/homie-api.service deploy/homie-kiosk.service dist/pkgroot/lib/systemd/system/
    install -m 0440 deploy/debian/sudoers dist/pkgroot/etc/sudoers.d/homie-deploy
    python3 deploy/blank-cursor.py dist/pkgroot/usr/share/homie/cursors default
    version="0.0.0+$(date -u +%Y%m%d%H%M%S).g$(git rev-parse --short HEAD)"
    size="$(du -sk dist/pkgroot | cut -f1)"
    mkdir -p dist/pkgroot/DEBIAN
    install -m 0755 deploy/debian/postinst deploy/debian/prerm deploy/debian/postrm dist/pkgroot/DEBIAN/
    sed -e "s/@VERSION@/$version/" -e "s/@INSTALLED_SIZE@/$size/" deploy/debian/control.in > dist/pkgroot/DEBIAN/control
    docker run --rm -v "$(pwd)/dist:/work" -w /work debian:trixie-slim dpkg-deb --build --root-owner-group pkgroot homie.deb

# Ship dist/homie.deb to the Pi and install it. The sudoers rule the package
# installs makes this non-interactive after the first, manual install (see
# README). The install triggers postinst, which restarts both services.
deploy: package
    rsync -az ./dist/homie.deb {{pi_host}}:/var/cache/homie/homie.deb
    ssh {{pi_host}} "sudo apt-get install -y /var/cache/homie/homie.deb"
    ssh {{pi_host}} "rm -f /var/cache/homie/homie.deb"

# Install WeatherKit credentials without putting the private key in the deploy tree.
deploy-weatherkit:
    #!/usr/bin/env bash
    set -euo pipefail
    test -f services/api/.env
    : "${WEATHERKIT_KEY_FILE:?set WEATHERKIT_KEY_FILE to the rotated WeatherKit .p8 file}"
    test -f "$WEATHERKIT_KEY_FILE"
    key_path="$(grep '^HOMIE_WEATHERKIT_KEY_PATH=' services/api/.env | cut -d= -f2-)"
    test -n "$key_path"
    ssh {{pi_host}} "sudo install -o homie -g homie -m 0700 -d /var/lib/homie"
    rsync -az "$WEATHERKIT_KEY_FILE" {{pi_host}}:/tmp/weatherkit.p8
    ssh {{pi_host}} "sudo install -o homie -g homie -m 0400 /tmp/weatherkit.p8 /var/lib/homie/weatherkit.p8 && rm /tmp/weatherkit.p8"
    scp services/api/.env {{pi_host}}:/tmp/homie-api.env
    ssh {{pi_host}} "sudo install -o root -g homie -m 0640 /tmp/homie-api.env {{pi_conf}}/api.env && rm /tmp/homie-api.env && sudo systemctl restart homie-api.service"

# Point the kiosk browser at this Mac's Vite dev server.
kiosk-dev:
    #!/usr/bin/env bash
    set -euo pipefail
    mac_host="$(scutil --get LocalHostName)"
    ssh {{pi_host}} "echo HOMIE_KIOSK_URL=http://${mac_host}.local:5173 > {{pi_conf}}/kiosk.env"
    ssh {{pi_host}} "sudo systemctl restart homie-kiosk.service"

# Point the kiosk browser back at the on-device API.
kiosk-prod:
    ssh {{pi_host}} "echo HOMIE_KIOSK_URL=http://localhost:3001 > {{pi_conf}}/kiosk.env"
    ssh {{pi_host}} "sudo systemctl restart homie-kiosk.service"
