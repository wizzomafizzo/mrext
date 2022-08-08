BINDIR       := $(CURDIR)/bin
MISTERBINDIR := $(BINDIR)/mister

ENVVARS := CGO_LDFLAGS="-lcurses"
LDFLAGS := -extldflags=-static

MISTERENVVARS      := GOOS="linux" GOARCH="arm" GOARM="7"
DOCKERIMAGENAME    := mister-mrext-build
DOCKERGOBUILDCACHE := mister-mrext-build-cache
DOCKERGOMODCACHE   := mister-mrext-mod-cache

all: build

clean:
	rm -rf $(BINDIR)

build:
	$(ENVVARS) go build -o "$(BINDIR)/search" --ldflags="$(LDFLAGS)" ./cmd/search
	go build -o "$(BINDIR)/random" ./cmd/random
	go build -o "$(BINDIR)/samindex" ./cmd/samindex

docker-build:
	$(ENVVARS) $(MISTERENVVARS) go build -o "$(MISTERBINDIR)/search.sh" --ldflags="$(LDFLAGS)" ./cmd/search
	$(MISTERENVVARS) go build -o "$(MISTERBINDIR)/random.sh" ./cmd/random
	$(MISTERENVVARS) go build -o "$(MISTERBINDIR)/samindex" ./cmd/samindex

build-mister:
	docker run --platform linux/arm/v7 -v $(DOCKERGOBUILDCACHE):/root/.cache/go-build -v $(DOCKERGOMODCACHE):/root/go/pkg/mod -v $(CURDIR):/build $(DOCKERIMAGENAME)

docker-image:
	docker build -t $(DOCKERIMAGENAME) scripts
