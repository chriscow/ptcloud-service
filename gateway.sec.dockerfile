
#
# This doesn't work yet. It is just a reminder to
# use the distroless container base image in the future
#

# Start by building the application.
FROM golang:1.13-buster as build

WORKDIR /go/src/app

# Assuming gateway exe has been built 
COPY . /go/src/app
COPY ./cmd/gateway/.secrets/. /go/bin/gateway/.secrets

RUN go get -d -v ./...
RUN go build -o /go/bin/gateway /go/src/app/cmd/gateway/.

# Now copy it into our distroless base image.
FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/. /
CMD ["/gateway"]