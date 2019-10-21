# Build image
FROM golang:1.12 AS build-env

WORKDIR /app
ENV GO111MODULE=on
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
COPY . .
RUN go test ./...
RUN go build

# Run image
FROM alpine:3.10
RUN apk add --no-cache ca-certificates
RUN echo "hosts: files dns" > /etc/nsswitch.conf
COPY --from=build-env /app/flyte .
COPY --from=build-env /app/flow/flow-schema.json .
COPY --from=build-env /app/swagger swagger
EXPOSE 8080

ENTRYPOINT ["./flyte"]

