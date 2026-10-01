set dotenv-load

run:
    docker compose up web --build

lint:
    golangci-lint run --fix

ya_test:
    go vet -vettool="$(pwd)/.tools/statictest" ./... && \
    go build -o cmd/shortener/shortener ./cmd/shortener/ && \
    ./shortenertest_v2-darwin-arm64 -test.v --test.run=^TestIteration1$ \
        -binary-path=cmd/shortener/shortener
test:
    go test ./...

sqlc-generate:
    go tool sqlc generate

create_migration name:
    go tool migrate create -ext sql -dir migrations -seq {{name}}

migrate_up:
    migrate -database "postgres://postgres:postgres@localhost:5432/shortener?sslmode=disable" -path migrations up

migrate_down:
    migrate -database "postgres://postgres:postgres@localhost:5432/shortener?sslmode=disable" -path migrations down
