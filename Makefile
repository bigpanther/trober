.PHONY: lint
lint: check-format
	go get golang.org/x/lint/golint
	go vet ./...
	golint -set_exit_status=1 ./...

.PHONY: check-format
check-format:
	@echo "Running gofmt..."
	$(eval unformatted=$(shell find . -name '*.go' | grep -v ./.git | grep -v vendor | xargs gofmt -s -l))
	$(if $(strip $(unformatted)),\
		$(error $(\n) Some files are ill formatted! Run: \
			$(foreach file,$(unformatted),$(\n)    gofmt -w -s $(file))$(\n)),\
		@echo All files are well formatted.\
	)
.PHONY: test-demo
test-demo:
	buffalo task db:demo_create db:demo_drop
.PHONY: test
test:
	buffalo test -coverprofile=coverage.txt -covermode=atomic -race ./...
API_VERSION = 0.1.0
.PHONY: gen
gen:
	# Modify the publish action if needed when changing this
	cd sdk
	npx @openapitools/openapi-generator-cli generate -i trober.yaml -c config.yaml -g dart-dio -o trober_sdk -p pubVersion=$(API_VERSION)
	cp LICENSE trober_sdk/LICENSE
	echo "See https://github.com/bigpanther/trober/releases/tag/v$(API_VERSION)" > trober_sdk/CHANGELOG.md
	cp README.md trober_sdk/README.md

.PHONY: publish
publish:	gen
	cd sdk/trober_sdk; dart pub publish --dry-run

.PHONY: pg
pg:
	psql -h localhost -U postgres

.PHONY: migrate
migrate:
	buffalo pop migrate


.PHONY: seed
seed:
	buffalo task db:seed
