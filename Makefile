EXISTING_VERSION := $(shell git describe --abbrev=0 --tags)
NEW_VERSION := $(shell echo $(EXISTING_VERSION) | awk -F. '{print ""$$1"."$$2"."$$3 + 1}')

.PHONY: tag_and_push test test-verbose test-coverage test-coverage-html bench test-race test-all

tag_and_push:
	git tag $(NEW_VERSION)
	git push origin $(NEW_VERSION)

# Test commands
test:
	go test -v

test-verbose:
	go test -v -cover

test-coverage:
	go test -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-coverage-html:
	go test -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

bench:
	go test -bench=. -benchmem

test-race:
	go test -race -v

test-all: test-verbose test-coverage bench test-race
	@echo "All tests completed successfully!"
