FROM docker.io/library/golang:1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN mkdir -p /app/dist
ARG VERSION=dev
ENV VERSION=${VERSION}
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X pichost.io/internal/config.version=${VERSION}" -modcacherw -o ./dist/ .;
FROM alpine:3.20 AS serve
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/dist/ /app/
ARG SPECS_VERSION=latest
ENV SPECS_VERSION=${SPECS_VERSION}
CMD ["/app/pichost.io", "http"]
