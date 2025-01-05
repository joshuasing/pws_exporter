# Copyright (c) 2024 Joshua Sing <joshua@joshuasing.dev>
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

# Build stage
FROM golang:1.23.4-alpine3.21@sha256:6c5c9590f169f77c8046e45c611d3b28fe477789acd8d3762d23d4744de69812 AS build

# Add ca-certificates, timezone data
RUN apk --no-cache add --update ca-certificates tzdata

# Create non-root user
RUN addgroup --gid 65532 pws_exporter && \
    adduser  --disabled-password --gecos "" \
    --home "/etc/pws_exporter" --shell="/sbin/nologin" \
    -G pws_exporter --uid 65532 pws_exporter

WORKDIR /build/pws_exporter
COPY . .

RUN go mod download
RUN go mod verify

RUN GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) CGO_ENABLED=0 GOGC=off \
    go build -trimpath -ldflags "-s -w" -o /build/pws_exporter/dist/pws_exporter ./cmd/pws_exporter

## Run stage
FROM scratch

# Build metadata
ARG VERSION
ARG VCS_REF
ARG BUILD_DATE

LABEL maintainer="Joshua Sing <joshua@joshuasing.dev>"
LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.authors="Joshua Sing <joshua@joshuasing.dev>" \
      org.opencontainers.image.url="https://github.com/joshuasing/pws_exporter" \
      org.opencontainers.image.source="https://github.com/joshuasing/pws_exporter" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Joshua Sing <joshua@joshuasing.dev>" \
      org.opencontainers.image.title="Personal Weather Station Exporter" \
      org.opencontainers.image.description="A Prometheus Exporter for off-the-shelf Personal Weather Stations (PWS)"

# Copy files
COPY --from=build /etc/group /etc/group
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/pws_exporter /etc/pws_exporter
COPY --from=build /build/pws_exporter/dist/pws_exporter /usr/local/bin/pws_exporter

USER pws_exporter:pws_exporter
ENTRYPOINT ["/usr/local/bin/pws_exporter"]
