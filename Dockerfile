# Build image
FROM golang:latest AS build-env

WORKDIR /app
ENV GO111MODULE=on
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
COPY . .
RUN go test ./...
RUN go build
ENTRYPOINT ["./flyte"]


# Run image
FROM alpine:latest
RUN echo "hosts: files dns" > /etc/nsswitch.conf
COPY --from=build-env /app/flyte .
COPY --from=build-env /app/flow/flow-schema.json .
COPY --from=build-env /app/swagger swagger
EXPOSE 8080

ENTRYPOINT ["./flyte"]

