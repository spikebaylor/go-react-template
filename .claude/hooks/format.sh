#!/usr/bin/env bash
# PostToolUse on Edit/Write: format the file just written, so formatting never
# costs a verify round trip. Errors are ignored: `just verify` is the gate.
set -u

file=$(node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{try{console.log(JSON.parse(s).tool_input.file_path??"")}catch{}})')
cd "$CLAUDE_PROJECT_DIR" || exit 0

case "$file" in
  *.go) gofmt -w "$file" ;;
  "$CLAUDE_PROJECT_DIR"/web/*) [ -d web/node_modules ] && npx --prefix web prettier --write --ignore-unknown "$file" ;;
esac >/dev/null 2>&1
exit 0
