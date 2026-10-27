# Builds the homie-api binary for the Pi (linux/arm64). This image is
# never run as a container on the device — it's a cross-compilation
# vehicle. Extract the binary with:
#   docker buildx build --platform linux/arm64 --target export --output dist .

FROM --platform=$BUILDPLATFORM node:22-bookworm AS frontend
WORKDIR /repo
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml turbo.json ./
COPY apps apps
COPY packages packages
RUN npm install -g pnpm@12.9.1 && pnpm install --frozen-lockfile
RUN pnpm exec turbo run build --filter=@homie/display --filter=@homie/manager

FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS api-builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
WORKDIR /src
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download
COPY services/api .
COPY --from=frontend /repo/apps/display/dist/ ./web/display/
COPY --from=frontend /repo/apps/manager/dist/ ./web/manager/
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT" -o /out/homie-api ./cmd/homie-api

FROM scratch AS export
COPY --from=api-builder /out/homie-api /homie-api
