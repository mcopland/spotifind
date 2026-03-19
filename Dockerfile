FROM golang:1.25-alpine AS go-builder
WORKDIR /server
COPY server/ .
RUN go build -o /app/spotifind ./cmd/spotifind

FROM node:20-alpine AS node-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM alpine:3.21
COPY --from=go-builder /app/spotifind /app/spotifind
COPY --from=node-builder /web/dist /app/dist
WORKDIR /app
CMD ["./spotifind"]
