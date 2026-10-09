FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY shared ./shared
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend .

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
COPY --from=build /out/backend /usr/local/bin/backend
ENV PORT=8080
EXPOSE 8080
USER app
ENTRYPOINT ["/usr/local/bin/backend"]
