.PHONY: phony
phony-goal: ; @echo $@

install: fetch-dependencies
	go install github.com/incu6us/goimports-reviser/v3@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/vladopajic/go-test-coverage/v2@latest
	go install github.com/kisielk/godepgraph@latest

fetch-dependencies:
	go mod download

generate: graph
	cd modules/common && go generate ./...
	cd modules/maths && go generate ./...
	cd modules/security && go generate ./...
	cd modules/servers && go generate ./...

graph:
	godepgraph -s ./modules/common | dot -Tpng -o ./docs/img/common.png
	godepgraph -s ./modules/maths | dot -Tpng -o ./docs/img/maths.png
	godepgraph -s ./modules/security | dot -Tpng -o ./docs/img/security.png
	godepgraph -s ./modules/servers | dot -Tpng -o ./docs/img/servers.png

imports:
	cd modules/common && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/maths && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/security && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/servers && goimports-reviser -rm-unused -set-alias -format -recursive .

format:
	cd modules/common && go fmt ./...
	cd modules/maths && go fmt ./...
	cd modules/security && go fmt ./...
	cd modules/servers && go fmt ./...

vet:
	cd modules/common && go vet ./...
	cd modules/maths && go vet ./...
	cd modules/security && go vet ./...
	cd modules/servers && go vet ./...

lint:
	cd modules/common && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...
	cd modules/maths && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...
	cd modules/security && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...
	cd modules/servers && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...

test:
	cd modules/common 	&& go test -covermode atomic -coverprofile .reports/coverage.out.tmp ./... 	&& cat .reports/coverage.out.tmp | grep -v "mocks.go" > .reports/coverage.out && rm .reports/coverage.out.tmp
	cd modules/maths 	&& go test -covermode atomic -coverprofile .reports/coverage.out.tmp ./... 	&& cat .reports/coverage.out.tmp | grep -v "mocks.go" > .reports/coverage.out && rm .reports/coverage.out.tmp
	cd modules/security && go test -covermode atomic -coverprofile .reports/coverage.out.tmp ./... 	&& cat .reports/coverage.out.tmp | grep -v "mocks.go" > .reports/coverage.out && rm .reports/coverage.out.tmp
	cd modules/servers 	&& go test -covermode atomic -coverprofile .reports/coverage.out.tmp ./... 	&& cat .reports/coverage.out.tmp | grep -v "mocks.go" > .reports/coverage.out && rm .reports/coverage.out.tmp

coverage: test
	cd modules/common 	&& go tool cover -func=.reports/coverage.out && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && go-test-coverage --config=.testcoverage.yml
	cd modules/maths 	&& go tool cover -func=.reports/coverage.out && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && go-test-coverage --config=.testcoverage.yml
	cd modules/security && go tool cover -func=.reports/coverage.out && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && go-test-coverage --config=.testcoverage.yml
	cd modules/servers 	&& go tool cover -func=.reports/coverage.out && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && go-test-coverage --config=.testcoverage.yml

check: fetch-dependencies imports format vet lint coverage

build: graph check

##

update-dependencies:
	cd modules/common 	&& go get -u ./... && go get -t -u ./... && go mod tidy
	cd modules/maths 	&& go get -u ./... && go get -t -u ./... && go mod tidy
	cd modules/security && go get -u ./... && go get -t -u ./... && go mod tidy
	cd modules/servers 	&& go get -u ./... && go get -t -u ./... && go mod tidy
	go work sync