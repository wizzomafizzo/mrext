FROM arm32v7/debian:bullseye

RUN echo "deb http://deb.debian.org/debian bullseye-backports main" >> /etc/apt/sources.list
RUN apt-get update && apt-get upgrade -y

RUN apt-get install -y ca-certificates openssl
ARG cert_location=/usr/local/share/ca-certificates
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
RUN update-ca-certificates

RUN apt-get install build-essential -y
RUN apt-get install -y golang-doc/bullseye-backports golang-go/bullseye-backports golang-src/bullseye-backports golang/bullseye-backports
RUN apt-get install -y sqlite3
RUN apt-get install libncurses-dev  -y

WORKDIR /build

CMD [ "make", "docker-build" ]