.PHONY: verify-tools install tidy graph generate imports format vet lint lint-inline build-inlineassign test bench coverage check validate build ci update-dependencies

MODULES := modules/compute/math modules/compute/engine modules/compute/tests/acceptance
MODULES += modules/config modules/core/common modules/core/crypto modules/core/security/authn modules/core/telemetry/otel modules/core/validation
MODULES += modules/extension/common/cache/redis modules/extension/common/cache/ristretto modules/extension/common/cast modules/extension/common/http/breaker modules/extension/common/http/limiter modules/extension/common/http/retry modules/extension/common/log/slog modules/extension/common/log/zerolog modules/extension/common/resilience/breaker modules/extension/common/resilience/limiter modules/extension/common/resilience/retry modules/extension/common/uids modules/extension/security/authn/grpc modules/extension/security/authn/http modules/extension/telemetry/otel/http modules/extension/telemetry/otel/slog
MODULES += modules/messaging
MODULES += modules/managed/cron modules/managed/diagnostics modules/managed/grpc modules/managed/http modules/managed/keep-alive
MODULES += sdks/decisions/core
ENABLE_INTERNAL := false
INTERNAL := internal/examples
INTERNAL += internal/temporal/courses/edu-101-go-code internal/temporal/courses/edu-102-go-code

# Modules that must be free of "No Inline Assignments" violations.
# Other modules (compute/math, compute/engine) still carry historical
# violations tracked under follow-up tickets; expand this list as those
# modules are cleaned up.
INLINE_MODULES := modules/core/common modules/core/validation modules/core/crypto

# Built inlineassign binary location. The cmd/inlineassign main package lives
# under tools/lint/inlineassign and is wired into go.work for local builds.
INLINEASSIGN_BIN := $(CURDIR)/tools/lint/inlineassign/bin/inlineassign

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
ifeq ($(ENABLE_INTERNAL),true)
	@for mod in $(INTERNAL); do \
		echo "==> imports $$mod"; \
        (cd $$mod && go tool goimports-reviser -rm-unused -set-alias -format -recursive .); \
	done
endif

format: verify-tools
	@for mod in $(MODULES); do \
		echo "==> format $$mod"; \
		(cd $$mod && go fmt ./...); \
	done
ifeq ($(ENABLE_INTERNAL),true)
	@for mod in $(INTERNAL); do \
		echo "==> format $$mod"; \
        (cd $$mod && go fmt ./...); \
	done
endif

vet: verify-tools
	@for mod in $(MODULES); do \
		echo "==> vet $$mod"; \
		(cd $$mod && go vet ./...); \
	done

lint: lint-inline
	@for mod in $(MODULES); do \
		echo "==> lint $$mod"; \
		(cd $$mod && go tool golangci-lint run --fix ./...); \
	done

build-inlineassign:
	@echo "==> build inlineassign analyzer"
	@mkdir -p $(dir $(INLINEASSIGN_BIN))
	@cd tools/lint/inlineassign && go build -o $(INLINEASSIGN_BIN) ./cmd/inlineassign

lint-inline: build-inlineassign
	@for mod in $(INLINE_MODULES); do \
		echo "==> lint-inline $$mod"; \
		(cd $$mod && go vet -vettool=$(INLINEASSIGN_BIN) ./...) || exit 1; \
	done

test: verify-tools
	@for mod in $(MODULES); do \
		echo "==> test $$mod"; \
		(cd $$mod && go test -covermode atomic -coverprofile .reports/testcoverage.out ./...); \
	done

# bench runs the per-algorithm crypto benchmarks under
# modules/core/crypto/*/examples/. CI does not invoke this target — the
# benchmark suite is opt-in to keep PR pipelines fast. Tune BENCHTIME to
# trade run time for accuracy (default 100ms).
BENCHTIME ?= 100ms
bench:
	@echo "==> bench modules/core/crypto (benchtime=$(BENCHTIME))"
	@cd modules/core/crypto && go test -bench=. -benchtime=$(BENCHTIME) -run=- ./...

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
