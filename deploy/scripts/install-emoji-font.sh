#!/bin/sh
# Installs Apple Color Emoji as a system font so the kiosk's emoji match the
# phone and the Mac. Chromium resolves "Apple Color Emoji" through fontconfig
# by name, and the display asks for that family first (--font-emoji), so
# dropping the TTF where fontconfig looks is all this has to do.
#
# The font is fetched at install time rather than shipped in homie.deb: the
# artwork is Apple's and is not ours to redistribute, and at 116 MB it would
# dwarf the rest of the package. Upstream publishes a .deb too, but dpkg holds
# its lock through our postinst, so a package cannot install another package —
# hence the bare TTF and a systemd unit outside the dpkg run.
#
# Run by homie-emoji-font.service, and retried by homie-emoji-font.timer until
# it succeeds: until then the kiosk shows whatever emoji font the Pi already
# had. Re-running once the font is in place is a no-op.
set -eu

# Pinned so a kiosk reinstalled next year renders what it renders today.
# To move to a newer macOS release, pick an asset from
# https://github.com/samuelngs/apple-emoji-ttf/releases and update both lines.
URL="https://github.com/samuelngs/apple-emoji-ttf/releases/download/macos-26-20260722-484daf4e/AppleColorEmoji-Linux.ttf"
SHA256="e37c7af6265ac4a0af6d57bc65e86109a776d9966e8343334557f63da482516f"

# /usr/local/share/fonts is the local-admin font directory: on fontconfig's
# default search path, and untouched by the distro's own font packages.
DEST=/usr/local/share/fonts/AppleColorEmoji.ttf
# The download lands in our own cache directory, not in the fonts directory,
# so a failed or half-finished fetch is never somewhere fontconfig scans.
WORK=/var/cache/homie/AppleColorEmoji.ttf.part

installed_sha() {
    [ -f "$1" ] || return 1
    sha256sum "$1" | cut -d' ' -f1
}

if [ "$(installed_sha "$DEST" || echo none)" = "$SHA256" ]; then
    echo "homie-emoji-font: $DEST is already current"
    exit 0
fi

mkdir -p "$(dirname "$DEST")" "$(dirname "$WORK")"
# Only the staged copy is cleaned up here. The partial download is kept on
# purpose so the next attempt resumes instead of re-fetching 116 MB.
trap 'rm -f "$DEST.new"' EXIT

echo "homie-emoji-font: fetching $URL"
# --http1.1 because HTTP/2 kept dying mid-transfer on this file with
# PROTOCOL_ERROR. --continue-at resumes a part-downloaded file, and
# --retry-all-errors covers the transfer errors that plain --retry ignores.
curl --fail --location --silent --show-error \
    --http1.1 \
    --continue-at - \
    --retry 5 --retry-delay 10 --retry-all-errors \
    --output "$WORK" "$URL"

got="$(installed_sha "$WORK")"
if [ "$got" != "$SHA256" ]; then
    echo "homie-emoji-font: checksum mismatch, discarding download" >&2
    echo "  expected $SHA256" >&2
    echo "  got      $got" >&2
    # A resumed download built on a corrupt partial would stay corrupt, so
    # drop it and let the next run start from nothing.
    rm -f "$WORK"
    exit 1
fi

# Copy in beside the target and rename, so DEST goes from absent to complete
# in one step even though the copy itself crosses filesystems.
install -m 0644 "$WORK" "$DEST.new"
mv "$DEST.new" "$DEST"
rm -f "$WORK"
fc-cache --force "$(dirname "$DEST")" >/dev/null

# Chromium reads fontconfig once at startup, so a running kiosk keeps drawing
# the old emoji until it restarts. try-restart leaves a stopped kiosk stopped.
systemctl try-restart homie-kiosk.service || true

echo "homie-emoji-font: installed $DEST"
