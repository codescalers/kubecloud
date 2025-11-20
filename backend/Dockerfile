FROM golang:1.25-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=1 go build -v -o ./myceliumcloud

FROM alpine:latest
COPY --from=builder /app/internal/infrastructure/mailservice/templates /app/internal/infrastructure/mailservice/templates
COPY --from=builder /app/myceliumcloud /myceliumcloud

EXPOSE 8080
CMD [ "/myceliumcloud" ]