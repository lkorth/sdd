#!/bin/sh
# Refresh the canvas of the records after Claude edits a file. If the project has no
# sdd/records.canvas, do nothing. If the refresh fails, exit 2, so that Claude reads the error.
canvas="${CLAUDE_PROJECT_DIR:-.}/sdd/records.canvas"
[ -f "$canvas" ] || exit 0
"${CLAUDE_PLUGIN_ROOT}/skills/sdd/bin/sdd" canvas --root "${CLAUDE_PROJECT_DIR:-.}" --out "$canvas" >/dev/null || exit 2
