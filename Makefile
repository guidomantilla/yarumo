.PHONY: phony

phony-goal: ; @echo $@

verify-tools:
	@go tool golangci-lint --version >/dev/null 2>&1 || { echo >&2 "golangci-lint is not installed. Run 'make install'"; exit 1; }
	@go tool goimports-reviser -version >/dev/null 2>&1 || { echo >&2 "goimports-reviser is not installed. Run 'make install'"; exit 1; }
	@go tool godepgraph --help >/dev/null 2>&1 || { echo >&2 "godepgraph is not installed. Run 'make install'"; exit 1; }
	@command -v dot >/dev/null 2>&1 || { echo >&2 "dot is not installed. Install it with your package manager. brew install graphviz - for macos"; exit 1; }
	@go tool go-test-coverage --version >/dev/null 2>&1 || { echo >&2 "go-test-coverage is not installed. Run 'make install'"; exit 1; }
	@go tool mockgen --version >/dev/null 2>&1 || { echo >&2 "go-test-coverage is not installed. Run 'make install'"; exit 1; }
	@go tool govulncheck --version >/dev/null 2>&1 || { echo >&2 "govulncheck is not installed. Run 'make install'"; exit 1; }
	@echo "All tools are installed."

install:
	cd tools && go get -tool github.com/incu6us/goimports-reviser/v3@latest
	cd tools && go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	cd tools && go get -tool go.uber.org/mock/mockgen@latest
	cd tools && go get -tool github.com/vladopajic/go-test-coverage/v2@latest
	cd tools && go get -tool github.com/kisielk/godepgraph@latest
	cd tools && go get -tool golang.org/x/vuln/cmd/govulncheck@latest
	cd tools && go mod tidy

tidy:
	cd tools && go mod tidy
	cd internal/deprecated && go mod tidy
	cd internal/examples && go mod tidy
	#cd internal/dlocal && go mod tidy
	cd modules/common && go mod tidy
	cd modules/maths && go mod tidy
	cd modules/telemetry/datadog && go mod tidy
	#cd modules/security && go mod tidy
	cd modules/servers && go mod tidy
	go work sync

graph: verify-tools
	go tool godepgraph -s ./modules/common | dot -Tpng -o ./docs/img/common.png
	go tool godepgraph -s ./modules/maths | dot -Tpng -o ./docs/img/maths.png
	go tool godepgraph -s ./modules/telemetry/datadog | dot -Tpng -o ./docs/img/datadog.png
	#go tool godepgraph -s ./modules/security | dot -Tpng -o ./docs/img/security.png
	#go tool godepgraph -s ./modules/servers | dot -Tpng -o ./docs/img/servers.png

generate: verify-tools graph
	cd modules/common && go generate ./...
	cd modules/maths && go generate ./...
	cd modules/telemetry/datadog && go generate ./...
	#cd modules/security && go generate ./...
	#cd modules/servers && go generate ./...

imports: verify-tools
	#cd internal/deprecated && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd internal/examples && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd internal/dlocal && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd modules/common && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/maths && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	cd modules/telemetry/datadog && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd modules/security && go tool goimports-reviser -rm-unused -set-alias -format -recursive .
	#cd modules/servers && go tool goimports-reviser -rm-unused -set-alias -format -recursive .

format: verify-tools
	#cd internal/deprecated && go fmt ./...
	#cd internal/examples && go fmt ./...
	#cd internal/dlocal && go fmt ./...
	#cd modules/common && go fmt ./...
	cd modules/maths && go fmt ./...
	cd modules/telemetry/datadog && go fmt ./...
	#cd modules/security && go fmt ./...
	#cd modules/servers && go fmt ./...

vet: verify-tools
	#cd modules/common && go vet ./...
	cd modules/maths && go vet ./...
	cd modules/telemetry/datadog && vet ./...
	#cd modules/security && go vet ./...
	#cd modules/servers && go vet ./...

lint:
	#cd modules/common && go tool golangci-lint run --fix ./...
	#cd modules/maths && go tool golangci-lint run --fix ./...
	cd modules/telemetry/datadog && go tool golangci-lint run --fix ./...
	#cd modules/security && go tool golangci-lint run --fix ./...
	#cd modules/servers && go tool golangci-lint run --fix ./...

test: verify-tools
	#cd modules/common && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...
	cd modules/maths && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...
	cd modules/telemetry/datadog && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...
	#cd modules/security && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...
	#cd modules/servers  && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...

coverage: verify-tools test
	#cd modules/common && go tool cover -html=.reports/testcoverage.out -o .reports/testcoverage.html && $(COVERAGE_BADGE) && go tool go-test-coverage --config=.testcoverage.yml || true
	cd modules/maths && go tool cover -html=.reports/testcoverage.out -o .reports/testcoverage.html && $(COVERAGE_BADGE) && go tool go-test-coverage --config=.testcoverage.yml || true
	cd modules/telemetry/datadog && go tool cover -html=.reports/testcoverage.out -o .reports/testcoverage.html && $(COVERAGE_BADGE) && go tool go-test-coverage --config=.testcoverage.yml || true
	#cd modules/security && go tool cover -html=.reports/testcoverage.out -o .reports/testcoverage.html && $(COVERAGE_BADGE) && go-test-coverage --config=.testcoverage.yml || true
	#cd modules/servers && go tool cover -html=.reports/testcoverage.out -o .reports/testcoverage.html && $(COVERAGE_BADGE) && go tool go-test-coverage --config=.testcoverage.yml || true

check: verify-tools
	cd modules/common && go tool govulncheck ./...
	cd modules/maths && go tool govulncheck ./...
	cd modules/telemetry/datadog && go tool govulncheck ./...
    #cd modules/security && go tool govulncheck ./...
    #cd modules/servers  && go tool govulncheck ./...

validate: verify-tools tidy generate imports format vet lint coverage check

build: validate

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