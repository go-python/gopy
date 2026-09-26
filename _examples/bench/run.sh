#!/bin/bash
# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Builds _examples/bench under each of gopy's backends and prints a table
# comparing run_bench.py's timing/memory numbers across them.  Run from the repo
# root; needs pybindgen, cffi and pybind11 all installed for the python
# interpreter named by $PYTHON (defaults to python3), and a C++ compiler.
set -eu

PYTHON="${PYTHON:-python3}"
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "building gopy..."
GOPY="$WORK/gopy"
(cd "$REPO" && go build -o "$GOPY" .)

printf '%-10s %8s %14s %14s %10s\n' backend calls "add (s)" "concat (s)" "peak (KB)"

for backend in pybindgen cffi pybind11; do
	out="$WORK/$backend"
	mkdir -p "$out"
	# gopy build cds into -output and runs go build there; give it a module
	# that resolves back to this checkout, same as go.mod already does for
	# anyone building _examples/* in place.
	printf 'module dummy\n\nrequire github.com/go-python/gopy v0.0.0\nreplace github.com/go-python/gopy => %s\n' "$REPO" >"$out/go.mod"
	GOPY_BACKEND="$backend" "$GOPY" build -vm="$PYTHON" -output="$out" -no-make -package-prefix= \
		"$REPO/_examples/bench" >"$out/build.log" 2>&1 || {
		echo "$backend: build failed, see $out/build.log" >&2
		tail -n 20 "$out/build.log" >&2
		continue
	}
	cp "$REPO/_examples/bench/run_bench.py" "$out/"
	row="$(cd "$out" && "$PYTHON" run_bench.py)"
	IFS=, read -r calls add_s concat_s peak_kb <<<"$row"
	printf '%-10s %8s %14s %14s %10s\n' "$backend" "$calls" "$add_s" "$concat_s" "$peak_kb"
done
