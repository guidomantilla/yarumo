.PHONY: verify-tools install tidy graph generate imports format vet lint test coverage check validate build ci update-dependencies

MODULES := modules/common modules/config modules/managed modules/telemetry/otel
MODULES += modules/maths modules/inference
ENABLE_INTERNAL := false
INTERNAL = internal/deprecated/passwords internal/deprecated/servers internal/dlocal internal/examples internal/maths
INTERNAL += internal/temporal/courses/edu-101-go-code internal/temporal/courses/edu-102-go-code

verify-tools:
	@go tool golangci-lint --version >/dev/null 2>&1 || { echo >&2 "golangci-lint is not installed. Run 'make install'"; exit 1; }
	@go tool goimports-reviser -version >/dev/null 2>&1 || { echo >&2 "goimports-reviser is not installed. Run 'make install'"; exit 1; }
	@go tool godepgraph --help >/dev/null 2>&1 || { echo >&2 "godepgraph is not installed. Run 'make install'"; exit 1; }
	@command -v dot >/dev/null 2>&1 || { echo >&2 "dot is not installed. Install it with your package manager. brew install graphviz - for macos"; exit 1; }
	@go tool go-test-coverage --version >/dev/null 2>&1 || { echo >&2 "go-test-coverage is not installed. Run 'make install'"; exit 1; }
	@go tool mockgen --version >/dev/null 2>&1 || { echo >&2 "mockgen is not installed. Run 'make install'"; exit 1; }
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
	@cd tools && go mod tidy
	@for mod in $(MODULES); do \
		echo "==> tidy $$mod"; \
		(cd $$mod && go mod tidy); \
	done
ifeq ($(ENABLE_INTERNAL),true)
	@for mod in $(INTERNAL); do \
		echo "==> tidy $$mod"; \
		(cd $$mod && go mod tidy); \
	done
endif
	@go work sync

graph: verify-tools
	@for mod in $(MODULES); do \
		name=$$(echo $$mod | sed 's|^[^/]*/||;s|/|-|g'); \
		echo "==> graph $$mod -> docs/img/$$name.png"; \
		go tool godepgraph -s ./$$mod | dot -Tpng -o ./docs/img/$$name.png; \
	done

generate: verify-tools graph
	@for mod in $(MODULES); do \
		echo "==> generate $$mod"; \
		(cd $$mod && go generate ./...); \
	done

imports: verify-tools
	@for mod in $(MODULES); do \
		echo "==> imports $$mod"; \
		(cd $$mod && go tool goimports-reviser -rm-unused -set-alias -format -recursive .); \
	done

format: verify-tools
	@for mod in $(MODULES); do \
		echo "==> format $$mod"; \
		(cd $$mod && go fmt ./...); \
	done

vet: verify-tools
	@for mod in $(MODULES); do \
		echo "==> vet $$mod"; \
		(cd $$mod && go vet ./...); \
	done

lint:
	@for mod in $(MODULES); do \
		echo "==> lint $$mod"; \
		(cd $$mod && go tool golangci-lint run --fix ./...); \
	done

test: verify-tools
	@for mod in $(MODULES); do \
		echo "==> test $$mod"; \
		(cd $$mod && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...); \
	done

coverage: verify-tools test
	@for mod in $(MODULES); do \
		echo "==> coverage $$mod"; \
		(cd $$mod && go tool cover -html=.reports/testcoverage.out -o .reports/testcoverage.html && $(COVERAGE_BADGE) && go tool go-test-coverage --config=.testcoverage.yml) || true; \
	done

check: verify-tools
	@for mod in $(MODULES); do \
		echo "==> govulncheck $$mod"; \
		output=$$(cd $$mod && go tool govulncheck ./... 2>&1); rc=$$?; \
		echo "$$output"; \
		if [ $$rc -eq 0 ]; then continue; fi; \
		if echo "$$output" | grep -q "^  Module:"; then \
			echo ""; echo "ERROR: third-party module vulnerabilities found in $$mod"; exit 1; \
		fi; \
		echo ""; echo "WARNING: only stdlib vulnerabilities found in $$mod (requires Go upgrade)"; \
	done

validate: verify-tools tidy generate imports format vet lint coverage

build: validate check

ci:
	@for mod in $(MODULES); do \
		echo "==> vet $$mod"; \
		(cd $$mod && go vet ./...); \
	done
	@for mod in $(MODULES); do \
		echo "==> lint $$mod"; \
		(cd $$mod && go tool golangci-lint run ./...); \
	done
	@for mod in $(MODULES); do \
		echo "==> test+coverage $$mod"; \
		(cd $$mod && go test -covermode atomic -coverprofile .reports/testcoverage.out ./... && go tool go-test-coverage --config=.testcoverage.yml) || true; \
	done

##
define COVERAGE_BADGE
COVERAGE_OUT=.reports/testcoverage.out COVERAGE_HTML=.reports/testcoverage.html; \
COVERAGE_PCT=$$(go tool cover -func="$$COVERAGE_OUT" | tail -1 | awk '{print $$3}') \
&& awk -v pct="$$COVERAGE_PCT" '{print} /<body[^>]*>/{print "<div style=\"position:fixed;top:10px;right:10px;padding:6px 10px;background:#222;color:#fff;border-radius:4px;font:14px/1.4 sans-serif;z-index:9999\">Total coverage: " pct "</div>"}' $$COVERAGE_HTML > $$COVERAGE_HTML.tmp \
&& mv $$COVERAGE_HTML.tmp $$COVERAGE_HTML
endef

##
update-dependencies:
	@for mod in $(MODULES); do \
		echo "==> update-dependencies $$mod"; \
		(cd $$mod && go get -u ./... && go get -t -u ./... && go mod tidy); \
	done
	@go work sync
