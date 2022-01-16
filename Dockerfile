FROM golang:1.17 as builder
WORKDIR /app
COPY ./ ./

ARG VERSION
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download
WORKDIR /app
RUN go version
RUN make build

FROM ubuntu:bionic
WORKDIR /app

# install CA certificates
RUN apt-get update && \
  apt-get install -y ca-certificates && \
  rm -Rf /var/lib/apt/lists/*  && \
  rm -Rf /usr/share/doc && rm -Rf /usr/share/man  && \
  apt-get clean

COPY --from=builder /app/.bin/changehub /app

RUN /app/changehub go-offline

ENTRYPOINT ["/app/changehub"]
