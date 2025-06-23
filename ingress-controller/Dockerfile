FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ingress-controller

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
COPY --from=builder /ingress-controller /ingress-controller
USER nobody
ENTRYPOINT ["/ingress-controller"]
