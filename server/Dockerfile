# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@v0.3.906

COPY *.go ./
COPY pkg ./pkg
COPY templates ./templates
RUN mkdir cmd
COPY cmd/manage ./cmd/manage
COPY cmd/migrate ./cmd/migrate

# Generate templates before building
RUN templ generate

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /process-api main.go
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /manage ./cmd/manage
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /migrate cmd/migrate/main.go

# Generate the agreement PDFs
FROM debian:12 AS build-pdf-stage

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tar xz-utils curl

RUN curl -L https://github.com/typst/typst/releases/download/v0.13.1/typst-x86_64-unknown-linux-musl.tar.xz \
    | tar --extract --xz --verbose --directory=/usr/bin --strip-components=1 typst-x86_64-unknown-linux-musl/typst

COPY pkg/resource/agreements/*.json /pkg/resource/agreements/
COPY pkg/resource/agreements/compileAgreementPDF.typ /pkg/resource/agreements/compileAgreementPDF.typ
COPY pkg/resource/fonts/ /pkg/resource/fonts/
COPY generate-agreement-pdf-files.sh /generate.sh

RUN bash /generate.sh

# Deploy the application binary into a lean image
# FROM gcr.io/distroless/base-debian12 AS build-release-stage
FROM debian:12 AS build-release-stage

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates

COPY --from=build-stage /process-api /process-api
COPY --from=build-stage /manage /manage
COPY --from=build-stage /migrate /migrate
COPY --from=build-pdf-stage /pkg/resource/agreements/*.json /pkg/resource/agreements/
COPY --from=build-pdf-stage /pkg/resource/agreements/*.pdf /pkg/resource/agreements/

COPY pkg/db/scripts /pkg/db/scripts
COPY pkg/db/seeds /pkg/db/seeds
COPY email-templates /email-templates
COPY static /static
# TODO: remove after consolidation
COPY fcmConfigs /fcmConfigs

EXPOSE 5000

# USER nonroot:nonroot

CMD ["/process-api"]
