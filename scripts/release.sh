#!/usr/bin/bash
#
# Release: next v1.<X> tag (patch segment ignored), push tags, warm proxy.golang.org.
#
# Run from the repository root (e.g. make release).

set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

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

printf 'git push --tags\n'
git push --tags

module=$(go list -m -q)
printf 'go list -m (warm GOPROXY) %s@%s\n' "$module" "$newtag"
# Resolve this version through the public proxy so it fetches the new tag from VCS.
GOPROXY=https://proxy.golang.org,direct go list -m "${module}@${newtag}"
