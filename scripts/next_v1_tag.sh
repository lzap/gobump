#!/usr/bin/bash
#
# Create the next release tag v1.<X>: take the largest X from existing tags
# matching v1.<digits>... (anything after the first number is ignored).
#
# Examples: v1.1.9 → X is 1; v1.10.0 → X is 10. New tag is v1.(max+1).

set -euo pipefail

mapfile -t tags < <(git tag -l 'v1.*')

latest=0
for t in "${tags[@]}"; do
	[[ -n "$t" ]] || continue
	[[ "$t" =~ ^v1\.([0-9]+)(\.|$) ]] || continue
	x="${BASH_REMATCH[1]}"
	((x > latest)) && latest=$x
done

next=$((latest + 1))
newtag="v1.$next"

printf 'git tag %s  (max v1.X seen: v1.%s)\n' "$newtag" "$latest"
git tag "$newtag"
