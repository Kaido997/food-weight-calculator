ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .


FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
COPY --from=builder /usr/src/app/internal/database/foodtable.json /internal/database/foodtable.json
COPY --from=builder /usr/src/app/internal/database/translations/ /internal/database/translations/
COPY --from=builder /usr/src/app/internal/database/analytics/ /internal/database/analytics/
CMD ["run-app"]
