FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go/go.mod go/go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
COPY go/ ./go/
RUN go build -o nile-auth ./go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/nile-auth /usr/local/bin/nile-auth
COPY go/migrations ./migrations
EXPOSE 8080
CMD ["/usr/local/bin/nile-auth"]
