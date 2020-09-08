COMMANDS = hubsync checkoutmaster maintenance prlist
REPOS ?= $(wildcard */)
OPTS ?= ;
REPOMAN ?= ~/go/src/moul.io/repoman

.PHONY: $(COMMANDS)
$(COMMANDS):
	@for repo in $(REPOS); do ( set -e; \
	  echo "cd $$repo && make -s -f $(REPOMAN)/Makefile _do.$@ $(OPTS)"; \
	  cd $$repo && make -s -f $(REPOMAN)/Makefile _do.$@ $(OPTS) \
	); done

_do.checkoutmaster: _do.hubsync
	git checkout master

_do.hubsync:
	hub sync

_do.prlist:
	@hub pr list -f "- %pC%>(8)%i%Creset %U - %t% l%n"

_do.maintenance: _do.checkoutmaster
	# renovate.json
	mkdir -p .github
	git mv renovate.json .github/renovate.json || true
	git rm -f renovate.json || true
	cp ~/go/src/moul.io/golang-repo-template/.github/renovate.json .github/ || true
	git add .github/renovate.json || true
	git add renovate.json || true

	# rules.mk
	if [ -f rules.mk ]; then cp ~/go/src/moul.io/rules.mk/rules.mk .; fi || true

	# authors
	if [ -f rules.mk ]; then make generate.authors; git add AUTHORS; fi || true

	# copyright
	sed -i "s/© 2014 /© 2014-2020 /" README.md
	sed -i "s/© 2015 /© 2015-2020 /" README.md
	sed -i "s/© 2016 /© 2016-2020 /" README.md
	sed -i "s/© 2017 /© 2017-2020 /" README.md
	sed -i "s/© 2018 /© 2018-2020 /" README.md
	sed -i "s/© 2019 /© 2019-2020 /" README.md

	# golangci-lint fix
	sed -i "s/version: v1.26/version: v1.28/" .github/workflows/*.yml || true
	sed -i "s/version: v1.27/version: v1.28/" .github/workflows/*.yml || true

	# apply changes
	git diff
	git diff --cached
	git branch -D dev/moul/maintenance || true
	git checkout -b dev/moul/maintenance
	git status
	git commit -s -a -m "chore: repo maintenance 🤖" -m "more details: https://github.com/moul/repoman"
	git push -u origin dev/moul/maintenance -f
	hub pull-request -m "chore: repo maintenance 🤖" -m "more details: https://github.com/moul/repoman" || $(MAKE) -f $(REPOMAN)/Makefile _do.prlist
