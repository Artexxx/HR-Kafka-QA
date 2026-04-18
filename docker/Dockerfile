FROM golang:1.26.2-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/main ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/bin/main ./main
COPY --from=builder /app/config/application.yaml ./config.yaml

RUN chown -R appuser:appgroup /home/appuser

USER appuser

EXPOSE 8080

ENTRYPOINT ["./main"]
CMD ["--config=./config.yaml"]
