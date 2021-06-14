FROM golang:1.13 AS build

COPY . /project
WORKDIR /project
RUN ls -la
RUN go build ./cmd/main.go

FROM ubuntu:20.04

RUN apt-get -y update && apt-get install -y tzdata

ENV PGVER 12
RUN apt-get -y update && apt-get install -y postgresql-$PGVER

USER postgres

RUN scripts/start.sh

EXPOSE 5432

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /usr/src/project

COPY . .
COPY --from=build /project/main .

EXPOSE 5000

CMD service postgresql start ./main