# El contexto de build es apps/backend. La etapa modules se comparte entre el
# entorno con recarga y la imagen final, evitando duplicar la resolución Go.
FROM golang:1.26.6-bookworm AS modules
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Air vive únicamente en esta etapa. Se resuelve desde el módulo de herramientas
# versionado; nunca se instala con @latest.
FROM modules AS dev
COPY go.tool.mod go.tool.sum ./
RUN go tool -modfile=go.tool.mod air -v
CMD ["go", "tool", "-modfile=go.tool.mod", "air", "-c", ".air.toml"]

FROM modules AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# scratch deja fuera compilador, fuentes y Air. El certificado permite las
# verificaciones TLS salientes que requiera la API.
FROM scratch AS runtime
COPY --from=modules /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/api /api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/api"]
