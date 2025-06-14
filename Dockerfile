FROM --platform=linux/x86_64 golang:latest AS base

WORKDIR /src

RUN apt-get update && \
    apt-get install -y git coreutils && \
    apt-get clean
COPY ./  .
ENV GOPROXY=https://proxy.golang.org
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go mod download

FROM base AS builder

ENV CGO_ENABLED=0
ARG VERSION_VAR

RUN mkdir /app
COPY ./  /app/
WORKDIR /app
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X \"main.version=${VERSION_VAR}\"" -o /out/ip-fetcher -- cmd/ip-fetcher/*.go

FROM --platform=linux/x86_64 gcr.io/distroless/static-debian12:nonroot
LABEL maintainer="Jon Hadfield jon@lessknown.co.uk"

WORKDIR /app
COPY publisher/README.template /app/publisher/README.template
COPY --from=builder /out/ip-fetcher /app/ip-fetcher
ENV GITHUB_TOKEN=$GITHUB_TOKEN
ENV GITHUB_PUBLISH_URL=$GITHUB_PUBLISH_URL
ENTRYPOINT ["/app/ip-fetcher"]
