# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/dubdutduc . \
	&& mkdir -p /out/data \
	&& chown 65532:65532 /out/data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build --chown=nonroot:nonroot /out/dubdutduc /app/dubdutduc
COPY --from=build --chown=nonroot:nonroot /out/data /data

VOLUME ["/data"]
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/dubdutduc"]
CMD ["-addr", ":8080", "-db", "/data/blocks.db"]
