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

ENV GATEWAY_ENDPOINT=:8080
ENV GCP_PROJECT_ID=strucim
ENV GCP_CREDENTIALS_PATH=./.secrets/strucim-gateway-keys.json
ENV IDENTIFY_BUCKET=pointcloud-identification
ENV INFERENCE_SERVICE=https://identify-shape-service-6wmkigtbzq-uc.a.run.app

WORKDIR /app
CMD ["./gateway"]