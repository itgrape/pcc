FROM node:22-alpine AS web-builder

WORKDIR /src/web

COPY web/package.json web/package-lock.json ./
RUN npm config set registry https://registry.npmmirror.com && npm ci --include=optional

COPY web/ ./

ARG PCC_CASDOOR_ENDPOINT
ARG PCC_CASDOOR_CLIENT_ID
ARG PCC_CASDOOR_APPLICATION=app-pcc
ARG PCC_CASDOOR_ORGANIZATION=built-in

RUN VITE_CASDOOR_URL="${PCC_CASDOOR_ENDPOINT}" \
    VITE_CASDOOR_CLIENT_ID="${PCC_CASDOOR_CLIENT_ID}" \
    VITE_CASDOOR_APP_NAME="${PCC_CASDOOR_APPLICATION}" \
    VITE_CASDOOR_ORG="${PCC_CASDOOR_ORGANIZATION}" \
    npm run build

FROM golang:1.25.5-alpine AS server-builder

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/pcc-server \
    ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S pcc \
    && adduser -S -G pcc pcc

WORKDIR /app

COPY --from=server-builder /out/pcc-server /usr/local/bin/pcc-server
COPY --from=web-builder /src/web/dist /web

EXPOSE 8080 8081

USER pcc

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz \
        && wget -q -O /dev/null http://127.0.0.1:8081/healthz \
        || exit 1

ENTRYPOINT ["/usr/local/bin/pcc-server"]
