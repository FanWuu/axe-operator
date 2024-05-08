#
# NOTE: THIS DOCKERFILE IS GENERATED VIA "apply-templates.sh"
#
# PLEASE DO NOT EDIT IT DIRECTLY.
#

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
	\
	mysqlsh --version

VOLUME /var/lib/mysql


EXPOSE 3306 33060
CMD [""]