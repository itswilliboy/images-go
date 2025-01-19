# build server app
FROM golang:1.22.10-alpine3.21 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

# build a fully standalone binary with zero dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o server .

RUN apk add --no-cache curl



# run server app
FROM alpine:3.21.2

# copy over SSL certificates, so that we can make HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /workspace/server /server

EXPOSE 3000

RUN apk add --no-cache curl

CMD ["/server"]
