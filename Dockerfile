# Build image
FROM golang:1.9-alpine AS build-env
ENV APP=flyte
RUN apk add --no-cache git curl
RUN git config --global http.sslVerify false
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep
WORKDIR $GOPATH/src/github.com/HotelsDotCom/$APP/
COPY . ./
RUN dep ensure -vendor-only
RUN go test ./...
RUN CGO_ENABLED=0 go build

# Run image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
RUN echo "hosts: files dns" > /etc/nsswitch.conf
ENV APP=flyte
WORKDIR /app
COPY --from=build-env /go/src/github.com/HotelsDotCom/$APP/$APP $APP
COPY --from=build-env /go/src/github.com/HotelsDotCom/$APP/swagger swagger
EXPOSE 8080
ENTRYPOINT "./$APP"