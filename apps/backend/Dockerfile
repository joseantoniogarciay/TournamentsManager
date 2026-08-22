# El tag 1.26.6-bookworm no está publicado todavía. Esta imagen 1.26.5 solo
# arranca el bootstrap: la directiva toolchain del módulo descarga y ejecuta
# Go 1.26.6 antes de resolver o compilar el backend.
FROM golang:1.26.5-bookworm AS modules
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go version && go mod download

# Air vive únicamente en esta etapa. Se resuelve desde el módulo de herramientas
# versionado; nunca se instala con @latest.
FROM modules AS dev
COPY go.tool.mod go.tool.sum ./
RUN go tool -modfile=go.tool.mod air -v
CMD ["go", "tool", "-modfile=go.tool.mod", "air", "-c", ".air.toml"]

FROM modules AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/legal-audit-backup ./cmd/legal-audit-backup
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/legal-audit-restore ./cmd/legal-audit-restore
RUN CGO_ENABLED=0 GOOS=linux go build -modfile=go.tool.mod -trimpath -ldflags="-s -w" -o /out/goose github.com/pressly/goose/v3/cmd/goose

# scratch deja fuera compilador, fuentes y Air. El certificado permite las
# verificaciones TLS salientes que requiera la API.
FROM scratch AS runtime
COPY --from=modules /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/api /api
COPY --from=build /out/legal-audit-backup /legal-audit-backup
COPY --from=build /out/legal-audit-restore /legal-audit-restore
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/api"]

# El migrador es una imagen separada: no se distribuye con la API ni se ejecuta
# en cada arranque. Solo recibe la credencial de migración en una ejecución
# explícita de Compose.
FROM scratch AS migrator
COPY --from=build /out/goose /goose
COPY db/migrations /migrations
ENTRYPOINT ["/goose"]
