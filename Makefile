.PHONY: phony
phony-goal: ; @echo $@

install: fetch-dependencies
	go install github.com/incu6us/goimports-reviser/v3@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/vladopajic/go-test-coverage/v2@latest

fetch-dependencies:
	go mod download

imports:
	goimports-reviser -rm-unused -set-alias -format -recursive internal
	goimports-reviser -rm-unused -set-alias -format -recursive pkg
	goimports-reviser -rm-unused -set-alias -format -recursive sdk
	go mod tidy

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

validate: check

##

update-dependencies:
	go get -u ./... && go get -t -u ./...
	go mod tidy