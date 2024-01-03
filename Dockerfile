FROM caddy:2-alpine

ARG caddyfile

COPY ${caddyfile} /etc/caddy/Caddyfile

RUN caddy validate --config /etc/caddy/Caddyfile

COPY assets assets
COPY files build/files
COPY build build
