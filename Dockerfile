# Build the manager binary
FROM golang:1.21 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/
COPY syncer/ syncer/
COPY sidecar/ sidecar/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details

FROM oraclelinux:8-slim

RUN set -eux; \
	groupadd --system --gid 999 mysql; \
	useradd --system --uid 999 --gid 999 --home-dir /var/lib/mysql --no-create-home mysql

# add gosu for easy step-down from root
# https://github.com/tianon/gosu/releases
ENV GOSU_VERSION 1.17
RUN set -eux; \
# TODO find a better userspace architecture detection method than querying the kernel
	arch="$(uname -m)"; \
	case "$arch" in \
		aarch64) gosuArch='arm64' ;; \
		x86_64) gosuArch='amd64' ;; \
		*) echo >&2 "error: unsupported architecture: '$arch'"; exit 1 ;; \
	esac; \
	curl -fL -o /usr/local/bin/gosu.asc "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$gosuArch.asc"; \
	curl -fL -o /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$gosuArch"; \
	export GNUPGHOME="$(mktemp -d)"; \
	gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4; \
	gpg --batch --verify /usr/local/bin/gosu.asc /usr/local/bin/gosu; \
	rm -rf "$GNUPGHOME" /usr/local/bin/gosu.asc; \
	chmod +x /usr/local/bin/gosu; \
	gosu --version; \
	gosu nobody true

RUN set -eux; \
	microdnf install -y \
		bzip2 \
		gzip \
		openssl \
		xz \
		zstd \
# Oracle Linux 8+ is very slim :)
		findutils \
	; \
	microdnf clean all


ENV MYSQL_MAJOR 8.0
ENV MYSQL_VERSION 8.0.36-1.el8

RUN set -eu; \
	{ \
		echo '[mysql8.0-server-minimal]'; \
		echo 'name=MySQL 8.0 Server Minimal'; \
		echo 'enabled=1'; \
		echo 'baseurl=https://repo.mysql.com/yum/mysql-8.0-community/docker/el/8/$basearch/'; \
		echo 'gpgcheck=0'; \
# https://github.com/docker-library/mysql/pull/680#issuecomment-825930524
		echo 'module_hotfixes=true'; \
	} | tee /etc/yum.repos.d/mysql-community-minimal.repo


RUN set -eu; \
	{ \
		echo '[mysql-tools-community]'; \
		echo 'name=MySQL Tools Community'; \
		echo 'baseurl=https://repo.mysql.com/yum/mysql-tools-community/el/8/$basearch/'; \
		echo 'enabled=1'; \
		echo 'gpgcheck=0'; \
# https://github.com/docker-library/mysql/pull/680#issuecomment-825930524
		echo 'module_hotfixes=true'; \
	} | tee /etc/yum.repos.d/mysql-community-tools.repo

ENV MYSQL_SHELL_VERSION 8.0.36-1.el8
RUN set -eux; \
	microdnf install -y "mysql-shell-$MYSQL_SHELL_VERSION"; \
	microdnf clean all; \
	mkdir //.mysqlsh && touch //.mysqlsh/mysqlsh.log;\
	chmod -R 666 //.mysqlsh/mysqlsh.log ; \
	mysqlsh --version

# FROM katanomi/distroless-static:nonroot

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
