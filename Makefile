MANSECTION ?= 1
SHELL=/bin/bash
VERSION := $(shell git describe --tags --abbrev=0)
.PHONY: clean snap
default: linux
all: linux windows darwin freebsd

linux: bin/linux-x86_64/mdview bin/linux-arm64/mdview bin/linux-i386/mdview
windows: bin/windows-x86_64/mdview.exe
darwin: bin/darwin-x86_64/mdview
freebsd: bin/freebsd-x86_64/mdview

deb: bin/linux-x86_64/mdview
	mkdir -p package/DEBIAN
	mkdir -p package/usr/bin/
	mkdir -p package/usr/share/man/man1/
	cp control package/DEBIAN/
	cp bin/linux-x86_64/mdview package/usr/bin/
	cp mdview.1 package/usr/share/man/man1/
	dpkg-deb --build package
	mv package.deb mdview-$(VERSION)_x86_64.deb

snap:
	snapcraft snap

bin/linux-x86_64/mdview: manpage
	env GOOS=linux GOARCH=amd64 go build -o ./bin/linux-x86_64/mdview-$(VERSION)/mdview
	cp mdview.1 bin/linux-x86_64/mdview-$(VERSION)/
	tar czvf mdview-$(VERSION)-linux-x86_64.tar.gz -C bin/linux-x86_64/ .

bin/linux-i386/mdview:
	env GOOS=linux GOARCH=386 go build -o ./bin/linux-i386/mdview-$(VERSION)/mdview
	cp mdview.1 bin/linux-i386/mdview-$(VERSION)/
	tar czvf mdview-$(VERSION)-linux-i386.tar.gz -C bin/linux-i386/ .

bin/linux-arm64/mdview:
	env GOOS=linux GOARCH=arm64 go build -o ./bin/linux-arm64/mdview-$(VERSION)/mdview
	cp mdview.1 bin/linux-arm64/mdview-$(VERSION)/
	tar czvf mdview-$(VERSION)-linux-arm64.tar.gz -C bin/linux-arm64/ .

bin/windows-x86_64/mdview.exe:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/windows-x86_64/mdview.exe
	zip -j mdview-$(VERSION)-windows-x86_64.zip bin/windows-x86_64/mdview.exe

bin/darwin-x86_64/mdview:
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin-x86_64/mdview-$(VERSION)/mdview
	cp mdview.1 bin/darwin-x86_64/mdview-$(VERSION)/
	tar czvf mdview-$(VERSION)-darwin-x86_64.tar.gz -C bin/darwin-x86_64/ .

bin/freebsd-x86_64/mdview:
	env GOOS=freebsd GOARCH=amd64 go build -o ./bin/freebsd-x86_64/mdview-$(VERSION)/mdview
	cp mdview.1 bin/freebsd-x86_64/mdview-$(VERSION)/
	tar czvf mdview-$(VERSION)-freebsd-x86_64.tar.gz -C bin/freebsd-x86_64/ .

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