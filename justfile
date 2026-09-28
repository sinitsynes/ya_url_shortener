set dotenv-load

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

sqlc-generate:
    go tool sqlc generate

create_migration name:
    migrate create -ext sql -dir migrations -seq {{name}}

migrate_up:
    migrate -database pgx5://postgres:postgres@localhost:5432/shortener -path migrations up

migrate_down:
    migrate -database pgx5://postgres:postgres@localhost:5432/shortener -path migrations down
