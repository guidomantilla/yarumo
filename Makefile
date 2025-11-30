.PHONY: phony
phony-goal: ; @echo $@

install:
	go install github.com/incu6us/goimports-reviser/v3@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/vladopajic/go-test-coverage/v2@latest
	go install github.com/kisielk/godepgraph@latest

tidy:
	cd internal/deprecated && go mod tidy
	cd internal/examples   && go mod tidy
	#cd internal/dlocal    && go mod tidy
	cd modules/common      && go mod tidy
	cd modules/maths       && go mod tidy
	#cd modules/security    && go mod tidy
	cd modules/servers     && go mod tidy
	go work sync

generate: graph
	cd modules/common   && go generate ./...
	cd modules/maths    && go generate ./...
	#cd modules/security && go generate ./...
	cd modules/servers  && go generate ./...

graph:
	godepgraph -s ./modules/common 	 | dot -Tpng -o ./docs/img/common.png
	godepgraph -s ./modules/maths 	 | dot -Tpng -o ./docs/img/maths.png
	#godepgraph -s ./modules/security | dot -Tpng -o ./docs/img/security.png
	godepgraph -s ./modules/servers  | dot -Tpng -o ./docs/img/servers.png

imports:
	cd internal/deprecated && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd internal/examples   && goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd internal/dlocal    && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/common      && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/maths       && goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd modules/security    && goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/servers     && goimports-reviser -rm-unused -set-alias -format -recursive .

format:
	cd internal/deprecated && go fmt ./...
	cd internal/examples   && go fmt ./...
	#cd internal/dlocal    && go fmt ./...
	cd modules/common      && go fmt ./...
	cd modules/maths       && go fmt ./...
	#cd modules/security    && go fmt ./...
	cd modules/servers     && go fmt ./...

vet:
	cd modules/common   && go vet ./...
	cd modules/maths    && go vet ./...
	#cd modules/security && go vet ./...
	cd modules/servers  && go vet ./...

lint:
	cd modules/common   && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...
	cd modules/maths    && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...
	#cd modules/security && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...
	cd modules/servers  && golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 ./...

test:
	cd modules/common   && go test -covermode atomic -coverprofile .reports/coverage.out ./...
	#cd modules/maths    && go test -covermode atomic -coverprofile .reports/coverage.out ./...
	#cd modules/security && go test -covermode atomic -coverprofile .reports/coverage.out ./...
	#cd modules/servers  && go test -covermode atomic -coverprofile .reports/coverage.out ./...

coverage: test
	cd modules/common   && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && $(COVERAGE_BADGE) && go-test-coverage --config=.testcoverage.yml || true
	#cd modules/maths    && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && $(COVERAGE_BADGE) && go-test-coverage --config=.testcoverage.yml || true
	#cd modules/security && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && $(COVERAGE_BADGE) && go-test-coverage --config=.testcoverage.yml || true
	#cd modules/servers  && go tool cover -html=.reports/coverage.out -o .reports/coverage.html && $(COVERAGE_BADGE) && go-test-coverage --config=.testcoverage.yml || true

check: tidy generate imports format vet lint coverage

build: check

##
define COVERAGE_BADGE
COVERAGE_OUT=.reports/coverage.out COVERAGE_HTML=.reports/coverage.html; \
COVERAGE_PCT=$$(go tool cover -func="$$COVERAGE_OUT" | tail -1 | awk '{print $$3}') \
&& awk -v pct="$$COVERAGE_PCT" '{print} /<body[^>]*>/{print "<div style=\"position:fixed;top:10px;right:10px;padding:6px 10px;background:#222;color:#fff;border-radius:4px;font:14px/1.4 sans-serif;z-index:9999\">Total coverage: " pct "</div>"}' $$COVERAGE_HTML > $$COVERAGE_HTML.tmp \
&& mv $$COVERAGE_HTML.tmp $$COVERAGE_HTML
endef

##
update-dependencies:
	cd modules/common 	&& go get -u ./... && go get -t -u ./... && go mod tidy
	cd modules/maths 	&& go get -u ./... && go get -t -u ./... && go mod tidy
	cd modules/security && go get -u ./... && go get -t -u ./... && go mod tidy
	cd modules/servers 	&& go get -u ./... && go get -t -u ./... && go mod tidy
	go work sync