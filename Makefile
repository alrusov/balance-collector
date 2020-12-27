COPYRIGHT = (C)_Alexey_Rusov_(rolic402@mail.ru),_2017-

OS   = linux
ARCH = amd64

APP = $(shell basename `pwd`)

BUILD_NUMBER_FILE = ../_BUILDS/$(APP)
ifneq ($(wildcard BUILD_NUMBER),)
	BUILD_NUMBER_FILE = BUILD_NUMBER
endif

ifneq ($(wildcard WITHOUT_GLOBAL_TAGS),)
	TAGS_FILES = $(wildcard TAGS)
else
	TAGS_FILES = $(wildcard ../_BUILDS/TAGS TAGS)
endif

SPACE =
SPACE +=

#-----------------------------------------------------------------------------------------------------------------------------------------------------#

TS   = $(shell date -u +'%Y-%m-%d %H:%M:%S')
YEAR = $(shell date -u +'%Y')

TS        := $(subst $(SPACE),_,$(strip $(TS)))
COPYRIGHT := $(subst $(SPACE),_,$(strip $(COPYRIGHT)$(YEAR)))

#-----------------------------------------------------------------------------------------------------------------------------------------------------#

VERSION =

ifneq ($(wildcard VERSION),)
	VERSION = $(shell cat VERSION)

	ifeq ($(wildcard $(BUILD_NUMBER_FILE)),)
		BUILD_NUMBER = 1
	else
		BUILD_NUMBER = $(shell cat $(BUILD_NUMBER_FILE))
	endif

	ifneq ($(strip $(BUILD_NUMBER)),)
		VERSION := $(VERSION).$(BUILD_NUMBER)
	endif
endif

#-----------------------------------------------------------------------------------------------------------------------------------------------------#

TAGS =

ifneq ($(TAGS_FILES),)
	TAGS = $(shell (cat $(TAGS_FILES) | tr -d '\r' | tr '\n' ' ' | sed 's/ *$$//' | tr ' ' '_'))
endif

#-----------------------------------------------------------------------------------------------------------------------------------------------------#

GO_FLAGS =
CGO      = 0
EXTRA_LD = -extldflags -static

ifeq ($(wildcard STATIC),)
	CGO      = 1
	EXTRA_LD =
endif

#-----------------------------------------------------------------------------------------------------------------------------------------------------#

.DEFAULT_GOAL = default

.PHONY: default pull

default:
	@echo Availavle targets: pull build


#-----------------------------------------------------------------------------------------------------------------------------------------------------#

pull:
	git pull

build:
	if [ -d "vendor/github.com" ]; then \
		echo 'Please delete the vendor/github.com directory'; \
		echo 'See https://github.com/golang/go/issues/19000 for details'; \
	else \
		CGO_ENABLED=$(CGO) \
		GOOS=$(OS) \
		GOARCH=$(ARCH) \
		go build -o "cmd/$(OS)/$(APP)" \
			$(GO_FLAGS) \
			--ldflags "$(EXTRA_LD) \
			-X github.com/alrusov/misc.appVersion=$(VERSION) \
			-X github.com/alrusov/misc.appTags=$(TAGS) \
			-X github.com/alrusov/misc.buildTime=$(TS) \
			-X github.com/alrusov/misc.copyright=$(COPYRIGHT)"; \
		\
		echo $$(($(BUILD_NUMBER) + 1)) >$(BUILD_NUMBER_FILE); \
	fi;

#-----------------------------------------------------------------------------------------------------------------------------------------------------#
