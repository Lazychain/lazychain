FROM golang:1.22-alpine as builder

USER root
WORKDIR /code

RUN apk add --no-cache ca-certificates build-base git libusb-dev linux-headers

ADD go.mod go.sum ./
RUN set -eux; \
  export ARCH=$(uname -m); \
  WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
  if [ ! -z "${WASM_VERSION}" ]; then \
  mkdir -p /code/downloads; \
  wget -O /code/downloads/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \
  fi; \
  cp /code/downloads/libwasmvm_muslc.a /usr/lib/libwasmvm_muslc.${ARCH}.a; \
  cp /code/downloads/libwasmvm_muslc.a /usr/lib/libwasmvm_muslc.a;

RUN go mod download;

COPY . .

RUN LEDGER_ENABLED=true BUILD_TAGS=muslc LINK_STATICALLY=true make build

FROM alpine:3.16

RUN apk --update add jq

COPY --from=builder /code/build/lazychaind /usr/bin/lazychaind

WORKDIR /opt

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657

ENTRYPOINT ["/bin/sh"]
