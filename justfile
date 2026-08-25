run:
    docker compose up web --build

lint:
    golangci-lint run --fix

ya_test:
    go build -o cmd/shortener/shortener ./cmd/shortener/ && \
    ./shortenertest_v2-darwin-arm64 -test.v -test.run=^TestIteration3$ -source-path=.

test:
    go test -v ./...
