COMMANDS = hubsync checkoutmaster
REPOS ?= $(wildcard */)
OPTS ?= ;

.PHONY: $(COMMANDS)
$(COMMANDS):
	for repo in $(REPOS); do ( set -xe; \
	  cd $$repo && make -f ../Makefile _do.$@ $(OPTS) \
	); done

_do.checkoutmaster: _do.hubsync
	git checkout master

_do.hubsync:
	hub sync

_do.bump_renovate: _do.checkoutmaster
	mkdir -p .github
	git mv renovate.json .github/renovate.json || true
	git rm -f renovate.json || true
	cp ~/go/src/moul.io/golang-repo-template/.github/renovate.json .github/
	git status
	git diff
	git checkout -b dev/moul/bump-renovate
	git commit renovate.json .github/renovate.json -m "chore: bump renovate"
	git push -u origin dev/moul/bump-renovate
	hub pull-request -m "chore: bump renovate"
