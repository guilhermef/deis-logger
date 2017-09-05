# deis-logger
# https://github.com/topfreegames/deis-logger
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

FROM golang:1.9-alpine3.6

MAINTAINER TFG Co <backend@tfgco.com>

RUN mkdir -p /app/bin

RUN apk update
RUN apk add git make musl-dev gcc

ADD . /go/src/github.com/topfreegames/deis-logger
RUN cd /go/src/github.com/topfreegames/deis-logger && \
  make build && \
  mv bin/deis-logger /app/deis-logger && \
  mv config /app/config && \
  mv Makefile /app/Makefile

WORKDIR /app

EXPOSE 8080

CMD /app/deis-logger start
