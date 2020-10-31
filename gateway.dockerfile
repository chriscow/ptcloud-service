FROM golang:alpine as builder

RUN mkdir /build

ADD . /build/
WORKDIR /build

RUN go build -o gateway ./cmd/gateway/.

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/gateway /app/
COPY --from=builder /build/cmd/gateway/.secrets/* /app/.secrets/

WORKDIR /app
CMD ["./gateway"]