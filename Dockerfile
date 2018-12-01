FROM golang:1.11 AS builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/kubebuild/pipelineme
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix nocgo -o app .

FROM alpine:latest
ENV GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
RUN apk --no-cache add ca-certificates git tar bash openssh-client
WORKDIR /root/
COPY --from=builder /go/src/github.com/kubebuild/pipelineme/app .
CMD ["./app"]