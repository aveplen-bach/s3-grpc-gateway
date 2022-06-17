FROM golang:1.18 as builder

RUN mkdir -p /go/src/github.com/aveplen-bach/s3

WORKDIR /go/src/github.com/aveplen-bach/s3

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -o /bin/s3imageapi /go/src/github.com/aveplen-bach/s3/cmd/s3

FROM alpine:3.15.4 as runtime

RUN apk add curl
COPY --from=builder /bin/s3imageapi /bin/s3imageapi

ENTRYPOINT [ "/bin/s3imageapi" ]