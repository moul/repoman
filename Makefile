COMMANDS = hubsync checkoutmaster
REPOS = $(wildcard */)

.PHONY: $(COMMANDS)
$(COMMANDS):
	for repo in $(REPOS); do ( set -xe; \
	  cd $$repo && make -f ../Makefile _do.$@; \
	); done

_do.checkoutmaster: _do.hubsync
	git checkout master

_do.hubsync:
	hub sync
