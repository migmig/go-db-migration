# syntax=docker/dockerfile:1.7

FROM node:22-bookworm AS frontend-build

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run verify:fast


FROM golang:1.25-bookworm AS backend-build

WORKDIR /src

RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-build /src/frontend/dist/ ./internal/web/assets/v16/

RUN CGO_ENABLED=1 GOOS=linux go build -tags nomsgpack -p=1 -o /out/dbmigrator .


FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app \
    && useradd --system --gid app --home-dir /data --create-home app

WORKDIR /data

ENV DBM_AUTH_DB_PATH=/data/auth.db

COPY --from=backend-build /out/dbmigrator /usr/local/bin/dbmigrator

USER app

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/dbmigrator"]
CMD ["-web"]
