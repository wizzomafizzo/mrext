# glibc in this version matches mister
FROM arm32v7/debian:bullseye-backports

RUN apt-get update -y && apt-get upgrade -y

# make go mod download work
RUN apt-get install -y ca-certificates openssl && \
    openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > /usr/local/share/ca-certificates/github.crt && \
    openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  /usr/local/share/ca-certificates/proxy.golang.crt && \
    update-ca-certificates

# install go and app dependencies
RUN apt-get install build-essential git -y && \
    apt-get install golang-doc/bullseye-backports golang-go/bullseye-backports golang-src/bullseye-backports golang/bullseye-backports -y && \
    apt-get install libsqlite3-dev libncurses-dev linux-source -y

# drop permissions on output files
RUN useradd -m -u 1000 build
USER build

# install mage
RUN cd /home/build && \
    git clone https://github.com/magefile/mage --depth 1 && \
    mkdir /home/build/go && \
    cd mage && go run bootstrap.go
ENV PATH="${PATH}:/home/build/go/bin"

WORKDIR /build