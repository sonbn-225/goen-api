# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS dev
WORKDIR /app

RUN apk add --no-cache ca-certificates git

# Install air for hot reload (requires Go >= 1.25)
RUN go install github.com/air-verse/air@v1.63.6

# Install swag CLI for Swagger docs generation (dev only)
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM golang:1.25-alpine AS build
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN mkdir -p /out/tmp
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -buildvcs=false -trimpath -tags "netgo,osusergo" -ldflags="-s -w" \
	-o /out/goen-api ./cmd/api

FROM scratch AS prod

WORKDIR /app

COPY --from=build /out/goen-api /app/goen-api
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /out/tmp /tmp

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/goen-api"]
