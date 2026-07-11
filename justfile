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

manpage:
	pandoc --standalone --to man mdview.1.md -o mdview.1