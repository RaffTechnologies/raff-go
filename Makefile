SPEC := ../docs/api-reference/openapi.yaml

.PHONY: generate verify build test vet lint sync clean

# Regenerate spec/spec.gen.go from the OpenAPI spec.
generate:
	go generate ./spec/...

# Regenerate and fail if anything changed (CI guard against drift).
verify: generate
	@if [ -n "$$(git status --porcelain spec/spec.gen.go)" ]; then \
		echo "ERROR: spec/spec.gen.go is out of sync with $(SPEC)."; \
		echo "Run 'make generate' and commit the result."; \
		git --no-pager diff spec/spec.gen.go; \
		exit 1; \
	fi

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed; skipping"

# Full local check — regenerate, build, vet, test.
sync: generate build vet test

clean:
	rm -f spec/spec.gen.go
