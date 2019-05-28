# Builder stage
FROM golang:alpine AS builder

# Install dependencies
RUN apk update && apk add --no-cache \
  git \
  ca-certificates \
  && update-ca-certificates

# Add source files and set the proper work dir
COPY .  $GOPATH/src/github.com/bysidecar/passport/
WORKDIR $GOPATH/src/github.com/bysidecar/passport/

# Fetch dependencies.
RUN go get -d -v
# Build the binary.
RUN go build -o /go/bin/passport

# Final image
FROM alpine
# Copy our static executable.
COPY --from=builder /go/bin/passport /go/bin/passport

# Copy the ca-certificates to be able to perform https requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Run the hello binary.
ENTRYPOINT ["/go/bin/passport"]
