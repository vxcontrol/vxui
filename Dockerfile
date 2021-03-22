# Dockerfile for publishing build to repo
FROM debian:buster-slim

RUN mkdir -p /opt/ui/logs

ADD app /opt/ui/bin/
ADD client/dist/ /opt/ui/static
ADD migrations /opt/ui/migrations
ADD templates /opt/ui/templates

WORKDIR /opt/ui

RUN apt update
RUN apt install -y ca-certificates
RUN apt clean

ENTRYPOINT ["/opt/ui/bin/app"]
