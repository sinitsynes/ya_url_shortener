run:
    docker compose up web --build

lint:
    golangci-lint run --fix

ya_test:
    go vet -vettool="$(pwd)/.tools/statictest" ./... && \
    go build -o cmd/shortener/shortener ./cmd/shortener/ && \
    ./shortenertest_v2-darwin-arm64 -test.v --test.run=^TestIteration6$ \
        -source-path=.
test:
    go test -cover ./...
