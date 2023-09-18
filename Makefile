MANSECTION ?= 1
SHELL=/bin/bash
.PHONY: clean snap
default: linux
all: linux windows darwin freebsd

linux: bin/linux-amd64/mdview bin/linux-arm64/mdview bin/linux-i386/mdview
windows: bin/windows-amd64/mdview.exe
darwin: bin/darwin-amd64/mdview bin/darwin-arm64/mdview
freebsd: bin/freebsd-amd64/mdview

deb: deb/linux-amd64 deb/linux-arm64

deb/linux-amd64: bin/linux-amd64/mdview
	mkdir -p package/DEBIAN
	mkdir -p package/usr/bin/
	mkdir -p package/usr/share/man/man1/
	cp control package/DEBIAN/
	sed -i "s/VERSION/$(VERSION)/g" package/DEBIAN/control
	cp -r bin/linux-amd64/mdview package/usr/bin/mdview
	cp mdview.1 package/usr/share/man/man1/
	dpkg-deb --build package
	mv package.deb mdview_$(VERSION)_amd64.deb

deb/linux-arm64: bin/linux-arm64/mdview
	mkdir -p package/DEBIAN
	mkdir -p package/usr/bin/
	mkdir -p package/usr/share/man/man1/
	cp control package/DEBIAN/
	sed -i "s/VERSION/$(VERSION)/g" package/DEBIAN/control
	cp -r bin/linux-arm64/mdview package/usr/bin/mdview
	cp mdview.1 package/usr/share/man/man1/
	dpkg-deb --build package
	mv package.deb mdview_$(VERSION)_arm64.deb

snap:
	snapcraft snap

bin/linux-amd64/mdview: manpage
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/linux-amd64/mdview
	cp mdview.1 bin/linux-amd64/
	tar czvf mdview-$(VERSION)-linux-amd64.tar.gz --transform s/linux-amd64/mdview-$(VERSION)/ -C bin linux-amd64

bin/linux-i386/mdview:
	env GOOS=linux GOARCH=386 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/linux-i386/mdview
	cp mdview.1 bin/linux-i386/
	tar czvf mdview-$(VERSION)-linux-i386.tar.gz --transform s/linux-i386/mdview-$(VERSION)/ -C bin linux-i386

bin/linux-arm64/mdview:
	env GOOS=linux GOARCH=arm64 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/linux-arm64/mdview
	cp mdview.1 bin/linux-arm64/
	tar czvf mdview-$(VERSION)-linux-arm64.tar.gz --transform s/linux-arm64/mdview-$(VERSION)/ -C bin linux-arm64

bin/windows-amd64/mdview.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/windows-amd64/mdview.exe
	zip -j mdview-$(VERSION)-windows-amd64.zip bin/windows-amd64/mdview.exe

bin/darwin-amd64/mdview:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/darwin-amd64/mdview
	cp mdview.1 bin/darwin-amd64/
	tar czvf mdview-$(VERSION)-darwin-amd64.tar.gz --transform s/darwin-amd64/mdview-$(VERSION)/ -C bin darwin-amd64

bin/darwin-arm64/mdview:
	env GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/darwin-arm64/mdview
	cp mdview.1 bin/darwin-arm64/
	tar czvf mdview-$(VERSION)-darwin-arm64.tar.gz --transform s/darwin-arm64/mdview-$(VERSION)/ -C bin darwin-arm64

bin/freebsd-amd64/mdview:
	env GOOS=freebsd GOARCH=amd64 go build -ldflags "-X main.appVersion=$(VERSION)" -o ./bin/freebsd-amd64/mdview
	cp mdview.1 bin/freebsd-amd64/mdview
	tar czvf mdview-$(VERSION)-freebsd-amd64.tar.gz --transform s/freebsd-amd64/mdview-$(VERSION)/ -C bin freebsd-amd64

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