FROM caddy:2.8.4-builder-alpine AS builder

RUN xcaddy build \
    --with github.com/Mirror0/caddy-langneg
FROM caddy:2.8.4-alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
