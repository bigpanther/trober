ARG TROBER_VERSION=dev
ARG TROBER_COMMIT=dev
FROM gobuffalo/buffalo:v0.16.18 as builder
ARG TROBER_VERSION
ARG TROBER_COMMIT
ENV GO111MODULE on
ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /src/trober
WORKDIR /src/trober
ENV CGO_ENABLED=0
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

ADD . .
RUN buffalo build -o /bin/trober --ldflags="-X github.com/bigpanther/trober/actions.version=${TROBER_VERSION} -X github.com/bigpanther/trober/actions.commit=${TROBER_COMMIT}"

FROM alpine
RUN apk add --no-cache bash ca-certificates

WORKDIR /bin/

COPY --from=builder /bin/trober .

ENV GO_ENV=production
# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

CMD /bin/trober migrate; /bin/trober
