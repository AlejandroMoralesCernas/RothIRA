# === Stage 1: build React ===
FROM node:18-alpine AS web-build
WORKDIR /web
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# === Stage 2: build Go (needs 1.23+ to satisfy your go.mod) ===
FROM golang:1.24-alpine AS go-build
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build either ./cmd (if present) or the module root
RUN sh -lc 'if [ -d ./cmd ]; then go build -o /server ./cmd; else go build -o /server .; fi'

# === Final image ===
FROM alpine:3.20
WORKDIR /app
COPY --from=go-build /server /app/server
# If your React tool outputs "build" (CRA) this is correct.
# If you use Vite, change /web/build to /web/dist.
COPY --from=web-build /web/build /app/frontend-build
EXPOSE 8080
ENTRYPOINT ["/app/server"]
