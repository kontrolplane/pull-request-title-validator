FROM golang:1.26.2 AS build

WORKDIR /action

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o pull-request-title-validator

FROM alpine:3.21

COPY --from=build /action/pull-request-title-validator /pull-request-title-validator

ENTRYPOINT ["/pull-request-title-validator"]
