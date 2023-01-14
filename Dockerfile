FROM alpine:3.9 AS build

ARG VERSION=0.88.0

ADD https://github.com/gohugoio/hugo/releases/download/v${VERSION}/hugo_${VERSION}_Linux-64bit.tar.gz /hugo.tar.gz
RUN tar -zxvf hugo.tar.gz

RUN apk add --no-cache git

COPY . /site
WORKDIR /site

RUN /hugo --minify --enableGitInfo

FROM nginx:1.15-alpine

WORKDIR /usr/share/nginx/html/

RUN rm -fr * .??*

COPY _docker/ /etc/nginx/conf.d/

RUN chmod 0644 /etc/nginx/conf.d/expires.inc
RUN chmod 0644 /etc/nginx/conf.d/default.conf

COPY --from=build /site/public /usr/share/nginx/html