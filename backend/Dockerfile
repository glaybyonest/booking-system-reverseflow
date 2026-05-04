FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/backend-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/backend-worker ./cmd/worker

FROM alpine:3.20

RUN adduser -D -H reserveflow && apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=build /out/backend-api /app/backend-api
COPY --from=build /out/backend-worker /app/backend-worker
COPY migrations /app/migrations

USER reserveflow
EXPOSE 8080
CMD ["/app/backend-api"]
