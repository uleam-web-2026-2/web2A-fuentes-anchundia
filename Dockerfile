# La API vigente esta en este modulo; las carpetas anteriores se conservan.
FROM golang:1.27 AS compilacion
WORKDIR /src
COPY Trabajo_autonomo_Lealcaf_hito2/go.mod Trabajo_autonomo_Lealcaf_hito2/go.sum ./
RUN go mod download
COPY Trabajo_autonomo_Lealcaf_hito2/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /servidor .

# Solo se entrega el ejecutable. La configuracion llega al arrancar.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=compilacion /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
