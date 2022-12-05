FROM registry.metroscales.io/errorbudget/golang-extras:1.17 as builder

WORKDIR /myproject
COPY grafana-controller/go.mod grafana-controller/go.sum /myproject/
RUN go mod download

ENV CGO_ENABLED 0
COPY grafana-controller .

RUN ./swagger.sh run
RUN go generate -run="stringer" ./...

ARG BUILD_MODE
RUN if [ "$BUILD_MODE" = "debug" ]; then \
    go build -gcflags "all=-N -l" -o app ./main/; \
    else \
    go build -ldflags '-s -w' -o app ./main/; \
    fi

RUN if [ "$BUILD_MODE" = "" ]; then \
    golangci-lint run -c .golangci.yml ./... && \
    ginkgo -tags unitTests -randomizeAllSpecs -failFast ./...; \
    fi

FROM registry.metroscales.io/errorbudget/node:16.7.0-bullseye-slim as swagger-converter
WORKDIR /swagger
COPY --from=builder /myproject/docs/swagger.json swagger.json
RUN if [ "$BUILD_MODE" = "" ]; then \
    npm install -g swagger2openapi && \
    swagger2openapi swagger.json > openapi.json; \
    fi

FROM registry.metroscales.io/errorbudget/debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates --no-install-recommends apt-utils && apt-get clean
EXPOSE 8000

COPY --from=builder /myproject/app /app
COPY --from=builder /myproject/resource/ /resource
COPY --from=swagger-converter /swagger/openapi.json /doc/openapi.json
COPY grafana_source.zip /doc/oma_grafana_source.zip
ENTRYPOINT [ "/app" ]
