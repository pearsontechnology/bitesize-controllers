FROM alpine:3.5

LABEL maintainer "martin.devlin@pearson.com"

ENV VAULT_VERSION=0.8.3
ENV VAULT_HTTPS_PORT=8243

COPY controller /
COPY /run.sh /usr/bin/run.sh
RUN chmod +x /usr/bin/run.sh
CMD ["/usr/bin/run.sh"]
