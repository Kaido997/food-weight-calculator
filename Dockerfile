ARG GO_VERSION=1.27.1
FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /usr/src/app
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /run-app .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /run-app /usr/local/bin/run-app
COPY internal/database/foodtable.json ./internal/database/foodtable.json
COPY internal/database/translations/ ./internal/database/translations/
CMD ["run-app"]
