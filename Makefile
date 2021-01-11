SHELL=/bin/bash -o pipefail
APP_NAME = learn
DOCKER_REPO = dxas90
BIN_PLATFORMS := go java node python

# This version-strategy uses git tags to set the version string
git_branch	   := $(shell git rev-parse --abbrev-ref HEAD)
git_tag		  := $(shell git describe --exact-match --abbrev=0 2>/dev/null || echo "")
commit_hash	  := $(shell git rev-parse --verify HEAD)
commit_timestamp := $(shell date --date="@$$(git show -s --format=%ct)" --utc +%FT%T)

VERSION		  := $(shell git describe --tags --always --dirty)
version_strategy := commit_hash
ifdef git_tag
	VERSION := $(git_tag)
	version_strategy := tag
else
	ifeq (,$(findstring $(git_branch),master HEAD))
		ifneq (,$(patsubst release-%,,$(git_branch)))
			VERSION := $(git_branch)
			version_strategy := branch
		endif
	endif
endif

.PHONY: version

container-%:
	@docker build -t $(DOCKER_REPO)/$(APP_NAME):$(firstword $(subst _, ,$*))-${VERSION} -f docker/$(firstword $(subst _, ,$*)).dockerfile .

push-%:
	@docker push $(DOCKER_REPO)/$(APP_NAME):$(firstword $(subst _, ,$*))-${VERSION}

all-container: $(addprefix container-, $(BIN_PLATFORMS))

all-push: $(addprefix push-, $(BIN_PLATFORMS))

version:
	@echo ::set-output name=version::$(VERSION)
	@echo ::set-output name=version_strategy::$(version_strategy)
	@echo ::set-output name=git_tag::$(git_tag)
	@echo ::set-output name=git_branch::$(git_branch)
	@echo ::set-output name=commit_hash::$(commit_hash)
	@echo ::set-output name=commit_timestamp::$(commit_timestamp)

