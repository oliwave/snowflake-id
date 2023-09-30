FROM alpine:3.10 as runtime

COPY ./compile /usr/local/bin/webhook
RUN chmod +x /usr/local/bin/webhook

ENTRYPOINT ["webhook"]