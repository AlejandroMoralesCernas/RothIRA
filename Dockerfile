# 1) Build frontend
FROM node:18 AS frontendbuilder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 2) Build backend
FROM golang:alpine AS builder
WORKDIR /builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/binaryfile ./cmd

# 3) Dev stage (Air)
FROM golang:alpine AS dev
WORKDIR /app
RUN apk add --no-cache git curl
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air"]

# 4) FINAL: prod last so it’s the default
FROM gcr.io/distroless/base-debian12 AS prod
WORKDIR /app
COPY --from=builder /app/binaryfile /app/binaryfile
COPY --from=frontendbuilder /app/build /app/frontend-build
CMD ["/app/binaryfile"]