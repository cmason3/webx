FROM docker.io/library/golang:1.25 AS BUILD

ENV CGO_ENABLED 0

RUN set -eux; \

git clone https://github.com/cmason3/webx.git; \
cd webx; go build -ldflags="-s -w" -trimpath main.go -o /webx

ENTRYPOINT [ "/webx", "-l", "0.0.0.0", "-p", "8081" ]

