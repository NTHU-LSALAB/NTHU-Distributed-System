PATH := $(CURDIR)/bin:$(PATH)

MODULES := video comment

DOCKER_COMPOSE := docker compose

####################################################################################################
### Automatically include components' extensions and ad-hoc rules (makefile.mk)
###
-include */makefile.mk

####################################################################################################
### rule for the generate command
###

define make-generate-rules

.PHONY: $1.generate

# to generate individual service, override the command defined in the docker-compose.yml file
$1.generate::
	$(DOCKER_COMPOSE) run --rm proto make $1.proto

endef
$(foreach module,$(MODULES),$(eval $(call make-generate-rules,$(module))))

.PHONY: generate
generate:
	$(DOCKER_COMPOSE) run --rm proto

define make-proto-rules

$1.proto:: bin/protoc-gen-go bin/protoc-gen-go-grpc
	protoc \
		-I . \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		./modules/$1/pb/*.proto

endef
$(foreach module,$(MODULES),$(eval $(call make-proto-rules,$(module))))

proto: bin/protoc-gen-go bin/protoc-gen-go-grpc $(addsuffix .proto,$(MODULES))

bin/protoc-gen-go: go.mod
	go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go

bin/protoc-gen-go-grpc: go.mod
	go build -o $@ google.golang.org/grpc/cmd/protoc-gen-go-grpc

####################################################################################################
### rule for the test command
###
