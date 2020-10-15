MANSECTION ?= 1
SHELL=/bin/bash
VERSION := $(shell git describe --tags --abbrev=0)
.PHONY: clean snap
default: linux
all: linux windows darwin freebsd

linux: bin/linux-amd64/mdview bin/linux-arm64/mdview bin/linux-i386/mdview
windows: bin/windows-amd64/mdview.exe
darwin: bin/darwin-amd64/mdview
freebsd: bin/freebsd-amd64/mdview

deb: bin/linux-amd64/mdview
	mkdir -p package/DEBIAN
	mkdir -p package/usr/bin/
	mkdir -p package/usr/share/man/man1/
	cp control package/DEBIAN/
	cp bin/linux-amd64/mdview package/usr/bin/
	cp mdview.1 package/usr/share/man/man1/
	dpkg-deb --build package
	mv package.deb mdview-$(VERSION)_amd64.deb

snap:
	snapcraft snap

bin/linux-amd64/mdview: manpage
	env GOOS=linux GOARCH=amd64 go build -o ./bin/linux-amd64/mdview
	tar czvf mdview-$(VERSION)-linux-amd64.tar.gz -C bin/linux-amd64/ mdview
bin/linux-i386/mdview:
	env GOOS=linux GOARCH=386 go build -o ./bin/linux-i386/mdview
	tar czvf mdview-$(VERSION)-linux-i386.tar.gz -C bin/linux-i386/ mdview
bin/linux-arm64/mdview:
	env GOOS=linux GOARCH=arm64 go build -o ./bin/linux-arm64/mdview
	tar czvf mdview-$(VERSION)-linux-arm64.tar.gz -C bin/linux-arm64/ mdview

bin/windows-amd64/mdview.exe:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/windows-amd64/mdview.exe
	zip -j mdview-$(VERSION)-windows-amd64.zip bin/windows-amd64/mdview.exe
bin/darwin-amd64/mdview:
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin-amd64/mdview
	tar czvf mdview-$(VERSION)-darwin-amd64.tar.gz -C bin/darwin-amd64/ mdview
bin/freebsd-amd64/mdview:
	env GOOS=freebsd GOARCH=amd64 go build -o ./bin/freebsd-amd64/mdview
	tar czvf mdview-$(VERSION)-freebsd-amd64.tar.gz -C bin/freebsd-amd64/ mdview

install:
	cp bin/snap/mdview $(DESTDIR)
clean:
	rm -rf bin
	rm -rf package
	rm -f *.tar.gz
	rm -f *.zip
	rm -f *.deb
	rm mdview.1

	# snapcraft clean mdview -s pull

manpage:
	pandoc --standalone --to man mdview.1.md -o mdview.1