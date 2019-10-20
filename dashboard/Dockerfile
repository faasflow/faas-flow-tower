FROM golang:1.10.4-alpine3.8 as build

WORKDIR /go/src/github.com/s8sg/faas-flow-tower/dashboard

COPY server.go .
COPY models.go .
COPY views.go .
COPY controller.go .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o handler . \
    && go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover

FROM openfaas/of-watchdog:0.5.3 as watchdog
FROM alpine

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/s8sg/faas-flow-tower/dashboard/handler  .
COPY assets     assets

RUN chown -R app /home/app

USER app

ENV fprocess="./handler"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"

CMD ["fwatchdog"]
