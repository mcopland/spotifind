FROM golang:1.25-alpine AS go-builder
WORKDIR /server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN go build -o /app/spotifind ./cmd/spotifind
RUN go build -o /app/migrate ./cmd/migrate

FROM node:20-alpine AS node-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /app/spotifind /app/spotifind
COPY --from=go-builder /app/migrate /app/migrate
COPY server/internal/database/migrations /app/internal/database/migrations
COPY --from=node-builder /web/dist /app/dist
WORKDIR /app
EXPOSE 8080
CMD ["./spotifind"]
