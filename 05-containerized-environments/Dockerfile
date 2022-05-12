FROM golang:1.18 AS builder

WORKDIR /

COPY main.go /

RUN go build -gcflags="all=-N -l" -o server main.go

FROM ubuntu:jammy

COPY --from=builder /server /

EXPOSE 12345/tcp

CMD ["/server"]
