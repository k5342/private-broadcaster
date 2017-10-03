# use multi-stage build, require docker 17.05 later

FROM golang:alpine3.6 as build-env

ADD . /go/src/private-broadcaster

WORKDIR /go/src/private-broadcaster

RUN apk update \
	&& apk add g++ musl-dev

RUN GOOS=linux GOARCH=amd64 go build app.go


FROM alpine:3.6

ARG NGINX_VERSION=nginx-1.13.5
ARG NGINX_RTMP_MODULE_VERSION=1.2.0
ARG SERVICE_DIR=/opt
ARG PACKAGES="gcc g++ pcre pcre-dev openssl-dev make"

RUN apk update \
	&& apk add ca-certificates \
	&& update-ca-certificates \
	&& apk add openssl

RUN wget https://nginx.org/download/$NGINX_VERSION.tar.gz \
	&& tar xvf $NGINX_VERSION.tar.gz

WORKDIR $NGINX_VERSION

RUN wget https://github.com/arut/nginx-rtmp-module/archive/v$NGINX_RTMP_MODULE_VERSION.tar.gz \
	&& tar xvf v$NGINX_RTMP_MODULE_VERSION.tar.gz

RUN apk add --no-cache $PACKAGES 

RUN ./configure \
		--with-debug \
		--prefix=/usr/share/nginx \
		--sbin-path=/usr/local/sbin/nginx \
		--conf-path=/etc/nginx/conf/nginx.conf \
		--pid-path=/var/run/nginx.pid \
		--http-log-path=/var/log/nginx/access.log \
		--error-log-path=/var/log/nginx/error.log \
		--add-module=./nginx-rtmp-module-$NGINX_RTMP_MODULE_VERSION \
	&& make -j2 \
	&& make install \
	&& make clean \
	&& apk del $PACKAGES

WORKDIR /

RUN rm -rf $NGINX_VERSION $NGINX_VERSION.tar.gz

RUN apk add --no-cache pcre ffmpeg

COPY --from=build-env /go/src/private-broadcaster/app /app/
ADD ./.env /app/
ADD ./templates /app/templates/
ADD ./nginx.conf /etc/nginx/conf/

# To run compiled golang binary on alpine linux
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /app
EXPOSE 1935 80 8080 443

CMD nginx && ./app
