FROM alpine

RUN apk update && apk upgrade

COPY ./gateway ./usr/sbin

ENTRYPOINT ["gateway"]