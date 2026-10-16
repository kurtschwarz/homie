#!/usr/bin/env python3
"""Generate a fully transparent Xcursor theme for the kiosk.

cage 0.3.1 has no flag to hide the pointer. It draws the compositor's own
cursor in the middle of the output as soon as a pointer device exists, and
Chromium only replaces that image once the pointer *enters* its surface --
which never happens on a kiosk nobody touches. So the page's `cursor: none`
cannot win, and XCURSOR_SIZE=1 does not help either: libXcursor picks the
closest image size a theme actually ships rather than scaling one down, so
the stock 24px arrow comes back.

A theme whose every cursor is a single transparent pixel sidesteps all of
that -- the compositor still draws a cursor, there is just nothing in it.

The theme has to be called "default". cage passes NULL as the theme name
and wlroots resolves that to "default", never reading XCURSOR_THEME -- so a
theme installed under any other name is simply never looked up, which looks
exactly like the fix not working.

Writes <dest>/<theme>/cursors/* and an index.theme. Point the compositor at
it with XCURSOR_PATH=<dest>.
"""

import os
import struct
import sys

# Xcursor chunk type for an image, per the file format's magic constants.
IMAGE_TYPE = 0xFFFD0002

# A theme advertises images at nominal sizes; libXcursor picks the closest
# one. Every size here is the same transparent pixel, so whatever a client
# asks for, it gets nothing.
NOMINAL_SIZES = (1, 12, 16, 24, 32, 48, 64, 96, 128)

# wlroots asks for "left_ptr"; clients ask for the rest. Any name missing
# from the theme can fall back to a visible inherited cursor, so this list
# is deliberately broad rather than minimal.
NAMES = (
    "default", "left_ptr", "arrow", "top_left_arrow", "right_ptr",
    "pointer", "hand", "hand1", "hand2", "pointing_hand",
    "text", "xterm", "ibeam", "crosshair", "cross", "tcross",
    "wait", "watch", "progress", "left_ptr_watch", "half-busy",
    "help", "question_arrow", "whats_this", "dnd-none", "dnd-copy",
    "move", "fleur", "all-scroll", "grab", "grabbing", "openhand",
    "not-allowed", "no-drop", "forbidden", "circle", "crossed_circle",
    "col-resize", "row-resize", "size_hor", "size_ver", "size_all",
    "sb_h_double_arrow", "sb_v_double_arrow", "h_double_arrow",
    "v_double_arrow", "left_side", "right_side", "top_side", "bottom_side",
    "top_left_corner", "top_right_corner", "bottom_left_corner",
    "bottom_right_corner", "nw-resize", "ne-resize", "sw-resize",
    "se-resize", "n-resize", "e-resize", "s-resize", "w-resize",
    "ew-resize", "ns-resize", "nesw-resize", "nwse-resize",
    "copy", "alias", "link", "context-menu", "cell", "vertical-text",
    "zoom-in", "zoom-out", "X_cursor", "none",
)


def blank_cursor() -> bytes:
    """One Xcursor file holding a 1x1 fully transparent image per size."""
    header = struct.pack("<4sIII", b"Xcur", 16, 0x00010000, len(NOMINAL_SIZES))

    # Each table entry points at its chunk, so chunks start after the table.
    offset = len(header) + 12 * len(NOMINAL_SIZES)
    table = b""
    chunks = b""

    for size in NOMINAL_SIZES:
        table += struct.pack("<III", IMAGE_TYPE, size, offset)
        # chunk header (36 bytes), then width*height ARGB pixels. The single
        # pixel is 0x00000000: transparent, and premultiplied alpha makes
        # that unambiguous.
        chunk = struct.pack(
            "<IIIIIIIII",
            36,          # header size
            IMAGE_TYPE,  # type
            size,        # subtype: nominal size
            1,           # image version
            1,           # width
            1,           # height
            0,           # xhot
            0,           # yhot
            0,           # delay (ms); irrelevant for a static cursor
        ) + struct.pack("<I", 0x00000000)
        chunks += chunk
        offset += len(chunk)

    return header + table + chunks


def main() -> int:
    dest = sys.argv[1] if len(sys.argv) > 1 else "/opt/homie/cursors"
    theme = sys.argv[2] if len(sys.argv) > 2 else "default"

    cursors = os.path.join(dest, theme, "cursors")
    os.makedirs(cursors, exist_ok=True)

    with open(os.path.join(dest, theme, "index.theme"), "w") as f:
        f.write("[Icon Theme]\nName={}\n".format(theme))

    payload = blank_cursor()
    for name in NAMES:
        with open(os.path.join(cursors, name), "wb") as f:
            f.write(payload)

    print("wrote {} cursors to {}".format(len(NAMES), cursors))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
