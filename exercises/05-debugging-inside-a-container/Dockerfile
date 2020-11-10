FROM golang:1.15 AS builder

WORKDIR /

COPY main.go /

RUN go build -o debugme main.go

FROM golang:1.15

COPY --from=builder /debugme /

EXPOSE 12345/tcp

CMD ["/debugme"]
