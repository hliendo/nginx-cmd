# Multi-stage build: nginx + Brotli en Alpine (estable)
FROM golang:1.24-alpine3.20 AS go-builder
RUN apk add --no-cache make
WORKDIR /app
COPY tools/nginx-cmd /app
RUN make build

FROM alpine:3.20 AS builder

RUN apk add --no-cache --update \
    build-base \
    pcre-dev \
    zlib-dev \
    openssl-dev \
    brotli-dev \
    git \
    mercurial \
    linux-headers \
    curl \
    wget

ENV NGINX_VERSION=1.26.0
ENV NGINX_BROTLI_COMMIT=master

WORKDIR /build

# Descargar nginx
RUN wget http://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz && \
    tar zxvf nginx-${NGINX_VERSION}.tar.gz

# Descargar módulo Brotli
RUN git clone --recursive https://github.com/google/ngx_brotli.git -b ${NGINX_BROTLI_COMMIT}

# Compilar nginx con Brotli
RUN cd nginx-${NGINX_VERSION} && \
    ./configure --prefix=/opt/nginx \
    --with-http_ssl_module \
    --with-http_v2_module \
    --add-module=../ngx_brotli && \
    make && make install

# Imagen final
FROM alpine:3.20
RUN apk add --no-cache pcre zlib openssl brotli ca-certificates curl certbot busybox-extras

COPY --from=builder /opt/nginx /opt/nginx
COPY datacfg/nginx/conf/nginx.conf /opt/nginx/conf/nginx.conf
# Copiar estructura modular
COPY datacfg/nginx/conf/conf.d /opt/nginx/conf/conf.d
RUN mkdir -p /var/log/nginx

# Copiar nginx-cmd
COPY --from=go-builder /app/nginx-cmd /usr/local/bin/nginx-cmd

# Entrypoint para cron + nginx
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENV PATH="/opt/nginx/sbin:/usr/local/bin:$PATH"

EXPOSE 80 443

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

ENTRYPOINT ["/entrypoint.sh"]
