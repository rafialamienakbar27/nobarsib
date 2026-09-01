# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

# Dependensi di-cache terpisah supaya perubahan kode tidak memicu unduh ulang.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api    ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 nobarsib
# Timestamp API memakai +07:00 (§8.1); zona waktu container harus WIB.
ENV TZ=Asia/Jakarta

WORKDIR /app
COPY --from=build /out/api    /app/api
COPY --from=build /out/worker /app/worker
COPY migrations /app/migrations
COPY seed       /app/seed

USER nobarsib
EXPOSE 8080
CMD ["/app/api"]
