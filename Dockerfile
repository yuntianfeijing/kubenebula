# Build the manager binary
FROM hub.xesv5.com/library/golang:1.12.7-proxy.io as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY constants/ constants/
COPY utils/ utils/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.10
RUN echo http://mirrors.aliyun.com/alpine/v3.10/main > /etc/apk/repositories; \
echo http://mirrors.aliyun.com/alpine/v3.10/community >> /etc/apk/repositories
RUN apk add bash curl
RUN apk update \
  && apk add tzdata \
  && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
  && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /
COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
