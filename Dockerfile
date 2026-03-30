FROM --platform=$BUILDPLATFORM alpine:latest AS certloader
RUN apk add --no-cache ca-certificates
RUN update-ca-certificates

FROM scratch
ARG TARGETPLATFORM
EXPOSE 8080

# Copy CA Certificates
COPY --from=certloader /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Needed for remote console
COPY --from=certloader /bin/sleep /bin/sleep

COPY ${TARGETPLATFORM}/groundskeeper /usr/bin/groundskeeper

CMD [ "/usr/bin/groundskeeper", "--help" ]