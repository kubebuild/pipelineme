FROM golang:latest
WORKDIR /go/src/github.com/kubebuild/pipelineme/
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
ENV GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
RUN apk --no-cache add ca-certificates git tar bash openssh-client
WORKDIR /root/
COPY --from=0 /go/src/github.com/kubebuild/pipelineme/app .
CMD ["./app"]