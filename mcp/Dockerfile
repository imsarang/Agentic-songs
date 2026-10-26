FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG BIN=spotify
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/${BIN}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/server /app/server
ENV SPOTIFY_DATA_DIR=/data \
    GOOGLE_DATA_DIR=/data \
    MCP_TRANSPORT=http \
    MCP_ADDR=:8080
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["/app/server"]
