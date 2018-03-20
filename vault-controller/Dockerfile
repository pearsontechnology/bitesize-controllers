FROM alpine:3.5

LABEL maintainer "martin.devlin@pearson.com"

RUN apk --update add bash curl

COPY controller /
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
