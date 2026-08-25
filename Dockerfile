FROM golang:1.26.7-alpine AS builder

WORKDIR /build

COPY go.mod ./

COPY . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o app cmd/shortener/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /build/app .

ENTRYPOINT ["./app -a=0.0.0.0:8000 -b=localhost:8000"]
