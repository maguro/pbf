#!/usr/bin/env bash
#
# bench-compare.sh compares the benchmarks of a base commit with the
# benchmarks of the working tree.
#
# Usage: bench-compare.sh BASE COUNT BENCHSTAT
#
# The script builds one test binary for each package that has benchmarks,
# for the base and for the working tree. It then runs the binaries COUNT
# times, one sample each time, and it alternates the base and the working
# tree. Odd rounds run the base first. Even rounds run the working tree
# first. Thus a slow drift of the machine, for example from heat, affects
# both sides by the same quantity. BENCHSTAT compares the two results.

set -euo pipefail

base_ref=$1
count=$2
benchstat=$3

root=$(git rev-parse --show-toplevel)
work="$root/.cache/bench"
base_tree="$work/base"

rm -rf "$work"
git -C "$root" worktree prune
mkdir -p "$work/bin"

git -C "$root" worktree add --quiet --detach "$base_tree" "$base_ref"
trap 'git -C "$root" worktree remove --force "$base_tree"' EXIT

pkgs=()
while IFS= read -r dir; do
	if grep -qs '^func Benchmark' "$dir"/*_test.go; then
		rel=${dir#"$root"}
		pkgs+=(".${rel}")
	fi
done < <(cd "$root" && go list -f '{{.Dir}}' ./...)

if [ ${#pkgs[@]} -eq 0 ]; then
	echo "bench-compare: no package has benchmarks" >&2
	exit 1
fi

for i in "${!pkgs[@]}"; do
	pkg=${pkgs[$i]}
	(cd "$base_tree" && go test -c -o "$work/bin/base-$i.test" "$pkg")
	(cd "$root" && go test -c -o "$work/bin/head-$i.test" "$pkg")
done

run() {
	local side=$1 tree=$2

	for i in "${!pkgs[@]}"; do
		(cd "$tree/${pkgs[$i]}" &&
			"$work/bin/$side-$i.test" -test.run '^$' -test.bench . \
				-test.benchmem -test.count 1) >>"$work/$side.txt"
	done
}

for round in $(seq 1 "$count"); do
	echo "bench-compare: round $round of $count" >&2

	if [ $((round % 2)) -eq 1 ]; then
		run base "$base_tree"
		run head "$root"
	else
		run head "$root"
		run base "$base_tree"
	fi
done

"$benchstat" base="$work/base.txt" head="$work/head.txt"
