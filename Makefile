binary = ocinfo

linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/$(binary)

windows:
	GOOS=windows GOARCH=amd64 go build -o ./bin/$(binary).exe

compile_all:
	echo "Compiling for Windows, Linux and MacOS"
	GOOS=windows GOARCH=amd64 go build -o ./bin/$(binary)_windows_amd64
	GOOS=linux GOARCH=amd64 go build -o ./bin/$(binary)_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -o ./bin/$(binary)_darwin_amd64

push:
	echo "TODO: Push binaries to somewhere under the rainbow"

all: compile_all push
