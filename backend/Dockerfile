FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY shared/go.mod shared/go.sum ./shared/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=from=gomodcache,target=/tmp/gomodcache \
    mkdir -p /go/pkg/mod/cache/download && cp -a /tmp/gomodcache/. /go/pkg/mod/cache/download/ && \
    GOPROXY=file:///go/pkg/mod/cache/download go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/backend /usr/local/bin/backend
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/backend"]
