run:
    docker compose up web --build

lint:
    golangci-lint run --fix

ya_test:
    go vet -vettool="$(pwd)/.tools/statictest" ./... && \
    go build -o cmd/shortener/shortener ./cmd/shortener/ && \
    ./shortenertest_v2-darwin-arm64 -test.v --test.run=^TestIteration8$ \
        -binary-path=cmd/shortener/shortener
test:
    go test -cover ./...
