FROM golang:1.21.3 AS build
WORKDIR /action
COPY . .
RUN CGO_ENABLED=0 go build -o pull-request-title-validation

FROM alpine:latest 
COPY --from=build /action/pull-request-title-validation /pull-request-title-validation
ENTRYPOINT ["/pull-request-title-validation"]