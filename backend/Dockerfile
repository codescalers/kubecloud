FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /myceliumcloud

EXPOSE 8080
CMD [ "/myceliumcloud" ]