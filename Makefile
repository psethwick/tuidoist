.PHONY: prepare
prepare: filter/filter_parser.go

.PHONY: install
install: prepare
	go install

.PHONY: build
build: prepare
	go build

update:
	go mod edit -replace="github.com/psethwick/todoist=github.com/psethwick/todoist@psethwick"

filter_parser.go: filter/filter_parser.y
	goyacc -o filter/filter_parser.go filter/filter_parser.y

test: prepare
	go test ./... -v
