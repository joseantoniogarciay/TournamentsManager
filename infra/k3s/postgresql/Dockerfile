# PostgreSQL 18 requires pgBackRest >= 2.55. Pin the same version validated in
# dev so the K3s image can archive WAL and create physical backups.
FROM debian:bookworm-slim AS pgbackrest-build

ARG PGBACKREST_VERSION=2.59.1

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        ca-certificates \
        curl \
        gcc \
        libbz2-dev \
        liblz4-dev \
        libpq-dev \
        libssh2-1-dev \
        libssl-dev \
        libxml2-dev \
        libzstd-dev \
        meson \
        ninja-build \
        pkg-config \
        zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

RUN curl --fail --location --silent --show-error \
        "https://github.com/pgbackrest/pgbackrest/releases/download/release%2F${PGBACKREST_VERSION}/pgbackrest-${PGBACKREST_VERSION}.tar.gz" \
        --output /tmp/pgbackrest.tar.gz \
    && mkdir /src \
    && tar --extract --gzip --file /tmp/pgbackrest.tar.gz --strip-components=1 --directory /src \
    && meson setup /build /src \
    && ninja -C /build

FROM postgres:18.4-bookworm

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        libbz2-1.0 \
        liblz4-1 \
        libssh2-1 \
        libxml2 \
        libzstd1 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=pgbackrest-build /build/src/pgbackrest /usr/local/bin/pgbackrest
COPY pgbackrest.conf /etc/pgbackrest/pgbackrest.conf

RUN chmod 0644 /etc/pgbackrest/pgbackrest.conf
