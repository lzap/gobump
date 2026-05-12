#!/usr/bin/bash
#
# Release: next semver tag v1.<minor>.0, push tags, warm proxy.golang.org.
#
# Go modules and the proxy require three-part versions (e.g. v1.2.0); v1.2 alone
# is not a valid module version. We bump <minor> and always use patch .0.
#
# Parsing: v1.MINOR.PATCH → MINOR; v1.MINOR (two segments) → MINOR for legacy tags.
#
# Run from the repository root (e.g. make release).

set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

mapfile -t tags < <(git tag -l 'v1.*')

max_minor=-1
for t in "${tags[@]}"; do
	[[ -n "$t" ]] || continue
	m=
	if [[ "$t" =~ ^v1\.([0-9]+)\.([0-9]+)$ ]]; then
		m="${BASH_REMATCH[1]}"
	elif [[ "$t" =~ ^v1\.([0-9]+)$ ]]; then
		m="${BASH_REMATCH[1]}"
	else
		continue
	fi
	((m > max_minor)) && max_minor=$m
done

if ((max_minor < 0)); then
	max_minor=0
fi

next=$((max_minor + 1))
newtag="v1.${next}.0"

printf 'git tag %s  (largest v1.<minor>.* seen: v1.%s.*)\n' "$newtag" "$max_minor"
git tag "$newtag"

printf 'git push --tags\n'
git push --tags

module=$(go list -m -f '{{.Path}}')
printf 'go list -m (warm GOPROXY) %s@%s\n' "$module" "$newtag"
GOPROXY=https://proxy.golang.org,direct go list -m "${module}@${newtag}"
