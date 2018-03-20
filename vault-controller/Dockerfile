FROM alpine:3.5

LABEL maintainer "martin.devlin@pearson.com"

RUN apk --update add bash

COPY controller /
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

CMD ["/usr/local/bin/docker-entrypoint.sh"]
