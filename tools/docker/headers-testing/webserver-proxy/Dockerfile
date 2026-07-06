FROM ubuntu:latest

COPY ./heavy_web_server /
RUN chmod +x /heavy_web_server
COPY ./entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
