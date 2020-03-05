.PHONY: all fmt build_sa build_ga clean

BINARY_TS=ts
BINARY_GA=ga

all: fmt build_ga build_ts build_all_ts build_all_ga

build_ga:
	go fmt ./genetic_algorithm/...
	go build -o ${BINARY_GA} ./genetic_algorithm/*.go

build_ts:
	go fmt ./tabu_search/...
	go build -o ${BINARY_TS} ./tabu_search/*.go


build_all_ts:
	go fmt ./tabu_search/...
	env GOOS=windows GOARCH=amd64 go build -o ${BINARY_TS}_Windows ./tabu_search/*.go
	env GOOS=darwin GOARCH=amd64 go build -o ${BINARY_TS}_MacOS ./tabu_search/*.go

build_all_ga:
	go fmt ./genetic_algorithm/...
	env GOOS=windows GOARCH=amd64 go build -o ${BINARY_GA}_Windows ./genetic_algorithm/*.go
	env GOOS=darwin GOARCH=amd64 go build -o ${BINARY_GA}_MacOS ./genetic_algorithm/*.go

clean:
	if [ -f ${BINARY_TS} ] ; then rm ${BINARY_TS} ; fi
	if [ -f ${BINARY_GA} ] ; then rm ${BINARY_GA} ; fi
	if [ -f ${BINARY_TS}_Windows ] ; then rm ${BINARY_TS}_Windows ; fi
	if [ -f ${BINARY_GA}_Windows ] ; then rm ${BINARY_GA}_Windows ; fi	
	if [ -f ${BINARY_TS}_MacOS ] ; then rm ${BINARY_TS}_MacOS ; fi
	if [ -f ${BINARY_GA}_MacOS ] ; then rm ${BINARY_GA}_MacOS ; fi

