default: all
all: linux windows darwin freebsd

linux: bin/linux-amd64/mdview
windows: bin/windows-amd64/mdview.exe
darwin: bin/darwin-amd64/mdview
freebsd: bin/freebsd-amd64/mdview

bin/linux-amd64/mdview:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/linux-amd64/mdview
bin/windows-amd64/mdview.exe:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/windows-amd64/mdview.exe
bin/darwin-amd64/mdview:
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin-amd64/mdview
bin/freebsd-amd64/mdview:
	env GOOS=freebsd GOARCH=amd64 go build -o ./bin/freebsd-amd64/mdview

.PHONY: clean
clean:
	rm -rf ./bin
