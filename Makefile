.PHONY: prepare
prepare: filter/filter_parser.go

.PHONY: install
install: prepare
	go install

.PHONY: build
build: prepare
	go build

filter_parser.go: filter/filter_parser.y
	goyacc -o filter/filter_parser.go filter/filter_parser.y

test: prepare
	go test ./... -v
