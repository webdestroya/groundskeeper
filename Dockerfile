FROM --platform=$BUILDPLATFORM alpine:latest AS certloader
RUN apk add --no-cache ca-certificates
RUN update-ca-certificates

# use busybox instead of scratch because we need some utilities for remote console
FROM busybox:latest

# Copy CA Certificates
COPY --from=certloader /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY ${TARGETPLATFORM}/groundskeeper /bin/groundskeeper

CMD [ "/bin/groundskeeper", "--help" ]