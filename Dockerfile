FROM golang:alpine as build
ENV CGO_ENABLED=0 GO_ENABLED=0

WORKDIR /src
COPY .. .

RUN apk add --no-cache git
RUN go mod tidy
RUN go build -ldflags "-s -w" -o ./bin/exapp ./main.go


FROM scratch as final

WORKDIR /app
COPY --from=build /src/bin/ /app/

EXPOSE 8080 9191

ENTRYPOINT ["/app/exapp"]