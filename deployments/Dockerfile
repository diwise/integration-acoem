FROM --platform=$BUILDPLATFORM docker.io/golang:1.20 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go test -v ./...

WORKDIR /app/cmd/integration-acoem

ARG TARGETARCH
ARG TARGETOS
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build


FROM registry.access.redhat.com/ubi9/ubi-minimal
WORKDIR /opt/diwise

LABEL org.opencontainers.image.source=https://github.com/diwise/integration-acoem

COPY --from=builder /app/cmd/integration-acoem/integration-acoem /opt/diwise/

RUN chmod 775 /opt/diwise

EXPOSE 8080
USER 1001

ENTRYPOINT ["/opt/diwise/integration-acoem"]
