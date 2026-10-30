# Copyright 2015-present, The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Reads the table run.sh prints on stdin, and prints its numbers as the JSON
# github-action-benchmark takes (tool: customSmallerIsBetter), for CI to keep
# from one run to the next: the time per call of Add and Concat, in
# microseconds, and the peak memory, for each backend that built.

from __future__ import print_function

import json
import sys

entries = []
for line in sys.stdin:
    fields = line.split()
    if len(fields) != 5:
        continue
    backend, calls, add_s, concat_s, peak_kb = fields
    try:
        calls, add_s, concat_s, peak_kb = int(calls), float(add_s), float(concat_s), int(peak_kb)
    except ValueError:
        continue  # the header
    for fn, seconds in (("Add", add_s), ("Concat", concat_s)):
        entries.append({
            "name": "%s: %s" % (backend, fn),
            "unit": "us/call",
            "value": round(seconds * 1e6 / calls, 4),
            "extra": "%d calls" % calls,
        })
    entries.append({"name": "%s: peak memory" % backend, "unit": "KB", "value": peak_kb})

json.dump(entries, sys.stdout, indent=2)
print()
