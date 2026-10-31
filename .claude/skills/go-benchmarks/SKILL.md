---
name: go-benchmarks
description: Measure and report the performance cost of a change. Use whenever a change touches a hot path that AGENTS.md lists, changes a default option value on such a path, or updates a dependency or the Go version, and whenever a change claims a performance gain.
---

# Measuring a change on a hot path

`AGENTS.md` lists the hot paths of this repo and the commands that measure them.

## What gets measured

A change that can alter the work done on a hot path is measured before and after
the change, on the same machine in the same session. This includes code on the
path, default option values on the path, and updates to dependencies on the path
and to the Go version.

When unsure whether a change qualifies, measure it. A change that does not alter
the work done on the success path, such as an error message, error wrapping, a
comment, or a rename, is not measured.

## How to measure

Alternate the base runs and the working-tree runs, one sample at a time, for at
least ten samples. Compare the results with `benchstat`. Use the compare command
that `AGENTS.md` names. If `AGENTS.md` names only a plain benchmark command,
check out the base in a separate worktree and alternate by hand.

Do not compare two runs made one after the other. Drift of the machine during
the first run makes a false difference in time.

If no benchmark covers the path that changed, add one in the same change. A
performance change is not done until the numbers show the gain. A correctness
change on a hot path reports its cost.

## Report the result in the commit message

The commit message states the result in a separate paragraph, for allocations
and bytes per op only, with no time figures. Allocation counts do not drift
between runs, and times age with every CPU and Go release.

- No change: `Benchmarks: no change (BenchmarkFoo, BenchmarkBar).`
- A change, where `benchstat` reports a significant difference and the values
  differ at three significant figures: before → after and the percentage, for
  each benchmark that changed. For example:
  `BenchmarkFoo: allocs/op 8.37M → 7.10M (−15.2%), B/op 1.90 GB → 1.62 GB (−14.7%).`

A change that is not measured has no benchmark paragraph.
