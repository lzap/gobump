#!/usr/bin/env python3
"""Create git tag v1.<X+1> where X is the max first segment of existing v1.* tags (patch ignored)."""

import re
import subprocess
import sys

PREFIX = "v1."


def main() -> int:
    out = subprocess.run(
        ["git", "tag", "-l", "v1.*"],
        check=True,
        capture_output=True,
        text=True,
    )
    tags = [t.strip() for t in out.stdout.splitlines() if t.strip()]

    max_x = 0
    pat = re.compile(r"^v1\.(\d+)(?:\.|$)")
    for t in tags:
        m = pat.match(t)
        if not m:
            continue
        max_x = max(max_x, int(m.group(1)))

    next_x = max_x + 1
    newtag = f"{PREFIX}{next_x}"
    print(f"git tag {newtag}  (max v1.X seen: {PREFIX}{max_x})")

    subprocess.run(["git", "tag", newtag], check=True)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except subprocess.CalledProcessError as e:
        print(e, file=sys.stderr)
        raise SystemExit(e.returncode or 1)
