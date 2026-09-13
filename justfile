run:
    docker compose up web --build

lint:
    golangci-lint run --fix

ya_test:
    go vet -vettool="$(pwd)/.tools/statictest" ./... && \
    go build -o cmd/shortener/shortener ./cmd/shortener/ && \
    ./shortenertest_v2-darwin-arm64 -test.v -test.run=^TestIteration5$ \
        -binary-path=cmd/shortener/shortener \
        -server-port=7999

test:
    go test -v ./...
