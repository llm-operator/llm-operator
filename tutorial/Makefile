.PHONE: default
default: sync

.PHONY: sync
sync: getting_started.md
	rm -f getting_started.ipynb
	jupytext --sync	getting_started.md
