# syntax=docker/dockerfile:1

FROM golang:1.25.5 AS build
WORKDIR /src
COPY go.mod ./
COPY main.go *.html ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/w42-eu-web .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/w42-eu-web /w42-eu-web
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/w42-eu-web", "-listen", ":8080"]
