FROM golang:latest AS builder

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/main.go

FROM alpine:latest AS runner

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app /app/app
WORKDIR /app

CMD ["./app"]
