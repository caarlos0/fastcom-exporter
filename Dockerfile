FROM alpine
EXPOSE 9877
COPY fastcom-exporter*.apk /tmp
RUN apk add --allow-untrusted /tmp/fastcom-exporter*.apk
ENTRYPOINT ["/usr/local/bin/fastcom-exporter"]
