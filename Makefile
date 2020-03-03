.PHONY: all fmt build_sa build_ga clean

BINARY_TS=ts
BINARY_GA=ga
BINARY_SA=sa

all: fmt build_sa build_ga build_ts

fmt:
	go fmt ./src/...
	go fmt ./genetic_algorithm/...

build_ga:
	go build -o ${BINARY_GA} ./genetic_algorithm/*.go

build_sa:
	go build -o ${BINARY_SA} ./simulated_annealing/*.go

build_ts:
	go build -o ${BINARY_TS} ./tabu_search/*.go

clean:
	if [ -f ${BINARY_TS} ] ; then rm ${BINARY_TS} ; fi
	if [ -f ${BINARY_GA} ] ; then rm ${BINARY_GA} ; fi
	if [ -f ${BINARY_SA} ] ; then rm ${BINARY_SA} ; fi
