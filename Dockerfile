# ==============================================================================
# flagcast — one parametrized Dockerfile for every Go service.
# Pick the service at build time:  --build-arg SVC=gateway
# Targets:  dev  (toolchain + source mount, used by compose)
#           prod (minimal non-root runtime)
# ==============================================================================

FROM golang:1.27-alpine AS base
RUN apk add --no-cache git build-base
WORKDIR /src
ENV CGO_ENABLED=0

FROM base AS dev
CMD ["go", "run", "./apps/gateway/cmd"]

FROM base AS build
ARG SVC=gateway
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -o /out/app ./apps/${SVC}/cmd

FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 app
COPY --from=build /out/app /usr/local/bin/app
USER app
ENTRYPOINT ["/usr/local/bin/app"]
