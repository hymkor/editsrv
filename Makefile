build :
	go fmt
	go build

package :
	zip -9 editsrv-%DATE:/=%.zip readme.md *.exe *.go *.cmd Makefile
