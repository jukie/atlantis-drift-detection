FROM --platform=$BUILDPLATFORM golang:1.20 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on go build -a -o drifter .

FROM --platform=$BUILDPLATFORM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/mutator .
USER nonroot:nonroot

ENTRYPOINT ["/drifter"]
