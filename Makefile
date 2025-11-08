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
	go generate ./modules/... ./sandbox/... ./sdk/... ./tools/...

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
	go fmt ./internal/... ./pkg/... ./sdk/... ./tools/...

vet:
	go vet ./internal/... ./pkg/... ./sdk/... ./tools/...

lint:
	golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./internal/... ./pkg/... ./sdk/...

test:
	go test -covermode atomic -coverprofile .reports/coverage.out.tmp ./internal/... ./pkg/... ./sdk/...
	cat .reports/coverage.out.tmp | grep -v "mocks.go" > .reports/coverage.out && rm .reports/coverage.out.tmp

coverage: test
	go tool cover -func=.reports/coverage.out
	go tool cover -html=.reports/coverage.out -o .reports/coverage.html
	go-test-coverage --config=.testcoverage.yml

check: fetch-dependencies imports format vet lint coverage

build: graph check

##

update-dependencies:
	go get -u ./... && go get -t -u ./...
	go mod tidy