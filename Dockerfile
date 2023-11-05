FROM golang:1.21

LABEL maintainer="David Oluwatobi"

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /usr/local/bin/app 

VOLUME ["/app/.env"]

EXPOSE 8080

CMD ["app", "--env-file=/app/.env"]