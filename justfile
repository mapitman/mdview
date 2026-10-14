set shell := ["/bin/bash", "-c"]

default: build

check:
	goreleaser check

build:
	goreleaser build --snapshot --clean

release-snapshot:
	goreleaser release --snapshot --clean --skip=publish

release:
	goreleaser release --clean

snap:
	snapcraft pack

clean:
	rm -rf dist
	rm -f mdview.1
	rm -f *.snap

	# snapcraft clean mdview -s pull

# The version defaults to the latest git tag; goreleaser passes its own.
manpage version=`git describe --tags --always`:
	sed 's|@VERSION@|{{version}}|g' mdview.1.md | pandoc --standalone --to man -o mdview.1
