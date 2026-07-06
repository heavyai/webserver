FROM goboring/golang:1.13.4b4

RUN mkdir /webserver
COPY ./src /webserver
RUN /webserver/scripts/build-all.sh
