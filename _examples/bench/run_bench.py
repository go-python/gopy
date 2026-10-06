# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Times N calls each of bench.Add and bench.Concat, and the peak memory
# after them, printing one CSV line: iterations,add_seconds,concat_seconds,peak_kb
# (see run.sh, which builds this example under each backend and prints the
# resulting rows as a table -- this script itself doesn't know which
# backend it was built with).

from __future__ import print_function

import sys
import time

import bench

if sys.platform == "win32":
    import psutil

    def peak_kb():
        return psutil.Process().memory_info().rss // 1024
else:
    import resource

    def peak_kb():
        return resource.getrusage(resource.RUSAGE_SELF).ru_maxrss


N = 1000
WARMUP = 100

for i in range(WARMUP):
    bench.Add(i, i)
    bench.Concat("a", "b")

start = time.perf_counter()
for i in range(N):
    bench.Add(i, i)
add_seconds = time.perf_counter() - start

start = time.perf_counter()
for i in range(N):
    bench.Concat("a", "b")
concat_seconds = time.perf_counter() - start

print("%d,%f,%f,%d" % (N, add_seconds, concat_seconds, peak_kb()))
