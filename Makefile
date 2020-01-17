.PHONY: all fmt build buildlinux clean

BINARY=main

all: fmt build

fmt:
	go fmt ./src/...

build: fmt
	go build -o ${BINARY} ./src/*.go

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

