FROM golang:alpine

#set working directory
WORKDIR /app

#install necessary modules to compile golang
COPY go.mod go.sum ./

COPY ./cmd ./cmd

RUN go mod download

WORKDIR /app/cmd

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]