.PHONY: all fmt build_sa build_ga clean

BINARY_TS_SA=ts_sa
BINARY_GA=ga

all: fmt build_sa build_ga

fmt:
	go fmt ./src/...
	go fmt ./genetic_algorithm/...

build_sa: fmt
	go build -o ${BINARY_TS_SA} ./src/*.go

clean:
	if [ -f ${BINARY_TS_SA} ] ; then rm ${BINARY_TS_SA} ; fi
	if [ -f ${BINARY_GA} ] ; then rm ${BINARY_GA} ; fi

build_ga:
	go build -o ${BINARY_GA} ./genetic_algorithm/*.go
