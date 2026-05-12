.PHONY: install tidy tag

install:
	go install ./...

tidy:
	go mod tidy

# Create git tag v1.<X> with X = 1 + max first segment of existing v1.* tags (patch ignored).
# Example: v1.1.9 and v1.1.2 → next v1.2; v1.10.0 → next v1.11
tag:
	@set -e; \
	latest=0; \
	for t in $$(git tag -l 'v1.*'); do \
		case $$t in v1.[0-9]*) ;; *) continue ;; esac; \
		x=$${t#v1.}; x=$${x%%.*}; \
		case $$x in ''|*[!0-9]*) continue ;; esac; \
		if [ "$$x" -gt $$latest ]; then latest=$$x; fi; \
	done; \
	next=$$((latest + 1)); \
	newtag=v1.$$next; \
	printf 'git tag %s  (max v1.X seen: v1.%s)\n' "$$newtag" "$$latest"; \
	git tag "$$newtag"
