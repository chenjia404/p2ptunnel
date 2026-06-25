FROM alpine:latest

# Keep the runtime image small while retaining CA certificates for update checks.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# GoReleaser dockers_v2 injects the selected platform artifact into the build context.
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/p2ptunnel /usr/local/bin/p2ptunnel
COPY LICENSE.txt /licenses/LICENSE.txt

# The default p2p port is 4001, and the forwarded service port defaults to 12000.
EXPOSE 4001/tcp
EXPOSE 4001/udp
EXPOSE 12000/tcp
EXPOSE 12000/udp

ENTRYPOINT ["/usr/local/bin/p2ptunnel"]
