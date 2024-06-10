FROM ghcr.io/gjermundgaraba/ignitecli:v28.3.0-go1.22 AS builder

USER root
WORKDIR /code

# Download dependencies and CosmWasm libwasmvm if found.
ADD go.mod go.sum ./
RUN --mount=type=cache,mode=0755,target=/code/downloads \
  set -eux; \
  export ARCH=$(uname -m); \
  WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
  if [ ! -z "${WASM_VERSION}" ]; then \
  wget -O /code/downloads/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \
  fi; \
  cp /code/downloads/libwasmvm_muslc.a /usr/lib/libwasmvm_muslc.${ARCH}.a; \
  cp /code/downloads/libwasmvm_muslc.a /usr/lib/libwasmvm_muslc.a;

RUN --mount=type=cache,mode=0755,target=/go/pkg/mod go mod download;

COPY . .

# Adds static linking to the build args in the ignite config.yml file
RUN printf "\n\
  ldflags:\n \
    - \"-linkmode=external\"\n \
    - \"-extldflags '-Wl,-z,muldefs -static'\"\n" >> config.yml

RUN --mount=type=cache,mode=0755,target=/root/.ignite ignite chain build --skip-proto --output build --build.tags muslc

FROM alpine:3.16

RUN apk --update add jq

COPY --from=builder /code/build/slothchaind /usr/bin/slothchaind

WORKDIR /opt

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657

ENTRYPOINT ["/bin/sh"]
