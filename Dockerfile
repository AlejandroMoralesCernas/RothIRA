FROM golang:alpine as builder
WORKDIR /builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd

FROM gcr.io/distroless/base-debian12 as prod
WORKDIR /app
COPY --from=builder /builder/main .
CMD ["./main"]

FROM golang:alpine as dev

#set working directory
WORKDIR /app

RUN apk add --no-cache git curl

RUN go install github.com/air-verse/air@latest 
#install necessary modules to compile golang
COPY go.mod go.sum ./

RUN go mod download

COPY . .

##RUN go build -o main ./cmd

EXPOSE 8080

CMD ["air"]