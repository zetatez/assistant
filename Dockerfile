# syntax=docker/dockerfile:1

FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY .env /app/.env
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
