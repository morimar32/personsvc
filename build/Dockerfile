FROM grpc_image:latest as grpc
COPY api/* /proto/
RUN ls /proto/* && cd / && /build_go.sh && ls /generated/*

# https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM morimar/go_build_base:latest AS builder

WORKDIR $GOPATH/src/github.com/morimar32/personsvc/

COPY . .
COPY --from=grpc /generated .
RUN mv ./generated/*.json ./web/static/swagger


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o svc; upx -9 -q svc
RUN cd cmd; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ping healthcheck.go; upx -9 -q ping; mv ping ../
RUN cp svc /
RUN cp ping /
RUN cp -R ./web/static/swagger /swagger
RUN cp .*.env /

FROM scratch
EXPOSE 8080 9090 6060
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /swagger /web/static/swagger
COPY --from=builder /svc /svc
COPY --from=builder /ping /ping
COPY --from=builder /.*.env /

USER appuser:appuser
ENTRYPOINT ["/svc"]
#HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/ping" ]