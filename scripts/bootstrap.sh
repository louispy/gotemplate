#!/usr/bin/env bash
set -euo pipefail

TEMPLATE_MODULE="github.com/louispy/gotemplate"
TEMPLATE_SLUG="gotemplate"

if [ "$#" -lt 1 ]; then
	echo "usage: $0 <new-module-path> [target-dir]" >&2
	echo "example: $0 github.com/<user>/<projectname>" >&2
	exit 1
fi

DST_MODULE="$1"
NAME="$(basename "$DST_MODULE")"
DST_DIR="${2:-$NAME}"

if ! command -v gonew >/dev/null 2>&1; then
	echo "gonew is not installed. install it with:" >&2
	echo "  go install golang.org/x/tools/cmd/gonew@latest" >&2
	exit 1
fi

if [ -e "$DST_DIR" ]; then
	echo "target '$DST_DIR' already exists" >&2
	exit 1
fi

gonew "$TEMPLATE_MODULE" "$DST_MODULE" "$DST_DIR"

cd "$DST_DIR"

for f in Makefile .env.example docker-compose.yml README.md; do
	if [ -f "$f" ]; then
		perl -i -pe "s/\\b${TEMPLATE_SLUG}\\b/${NAME}/g" "$f"
	fi
done

rm -rf scripts

git init -q -b main
git add -A

echo "created ${DST_DIR} (module ${DST_MODULE})"
echo "next:"
echo "  cd ${DST_DIR}"
echo "  go build ./..."
echo "  git commit -m 'Initial commit'"
