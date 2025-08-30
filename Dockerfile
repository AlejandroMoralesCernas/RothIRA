# 1) Build frontend
FROM node:18 AS frontendbuilder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 2) Build backend
FROM golang:alpine AS backendbuilder
WORKDIR /builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /builder/binaryfile ./cmd

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

# 4) frontend runtime
FROM nginx:alpine AS frontend
COPY --from=frontendbuilder /app/build /usr/share/nginx/html
EXPOSE 80

# 5) backend runtime
FROM gcr.io/distroless/base-debian12 AS backend
WORKDIR /app
COPY --from=backendbuilder /builder/binaryfile .
EXPOSE 8081
CMD ["./binaryfile"]