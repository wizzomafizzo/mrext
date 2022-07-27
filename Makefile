BINDIR       := $(CURDIR)/bin
MISTERBINDIR := $(BINDIR)/mister

ENVVARS := CGO_LDFLAGS="-lcurses"
LDFLAGS := -extldflags=-static

DOCKERIMAGENAME    := mister-mrext-build
DOCKERENVVARS      := GOOS="linux" GOARCH="arm" GOARM="7"
DOCKERGOBUILDCACHE := mister-mrext-build-cache
DOCKERGOMODCACHE   := mister-mrext-mod-cache

build:
	$(ENVVARS) go build -o "$(BINDIR)/search" --ldflags="$(LDFLAGS)" ./cmd/search
	$(ENVVARS) go build -o "$(BINDIR)/random" --ldflags="$(LDFLAGS)" ./cmd/random
	$(ENVVARS) go build -o "$(BINDIR)/samindex" --ldflags="$(LDFLAGS)" ./cmd/samindex

docker-build:
	$(ENVVARS) $(DOCKERENVVARS) go build -o "$(MISTERBINDIR)/search.sh" --ldflags="$(LDFLAGS)" ./cmd/search
#	$(ENVVARS) $(DOCKERENVVARS) go build -o "$(MISTERBINDIR)/samindex" --ldflags="$(LDFLAGS)" ./cmd/samindex

build-mister:
	docker run --platform linux/arm/v7 -v $(DOCKERGOBUILDCACHE):/root/.cache/go-build -v $(DOCKERGOMODCACHE):/root/go/pkg/mod -v $(CURDIR):/build $(DOCKERIMAGENAME)

docker-image:
	docker build -t $(DOCKERIMAGENAME) scripts
