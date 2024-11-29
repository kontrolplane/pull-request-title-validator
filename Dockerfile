FROM golang:1.23.1 AS build
WORKDIR /action
COPY . .
RUN CGO_ENABLED=0 go build -o pull-request-title-validator

FROM alpine:latest 
COPY --from=build /action/pull-request-title-validator /pull-request-title-validator
ENTRYPOINT ["/pull-request-title-validator"]