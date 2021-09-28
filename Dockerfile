FROM golang:alpine as builder

# package update
RUN apk update && \
    apk add --no-cache git mercurial

# コピー
WORKDIR /build
COPY . /build/

# 環境変数設定
ENV GO111MODULE=on
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Goビルド
# RUN go mod vendor
RUN go mod download
RUN go build -a -o goapp

# ビルド
FROM alpine:latest as production

# 環境変数設定
ENV PORT 8080

RUN apk --no-cache add tzdata ca-certificates
COPY --from=builder /build/goapp /goapp

EXPOSE 8080

CMD ["/goapp", "--host", "0.0.0.0"]
