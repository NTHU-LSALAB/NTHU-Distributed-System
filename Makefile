PATH := $(CURDIR)/bin:$(PATH)

MODULES := video comment

DOCKER_COMPOSE := $(or $(DOCKER_COMPOSE),$(DOCKER_COMPOSE),docker compose)

####################################################################################################
### Automatically include components' extensions and ad-hoc rules (makefile.mk)
###
-include */makefile.mk

.PHONY: clean
clean:
	rm -rf bin/*

####################################################################################################
### Rule for the `generate` command
###

define make-dc-generate-rules

.PHONY: dc.$1.generate

# to generate individual module, override the command defined in the docker-compose.yml file
dc.$1.generate:
	$(DOCKER_COMPOSE) run --rm generate make $1.generate

endef
$(foreach module,$(MODULES),$(eval $(call make-dc-generate-rules,$(module))))

.PHONY: dc.pkg.generate
dc.pkg.generate:
	$(DOCKER_COMPOSE) run --rm generate make pkg.generate

.PHONY: dc.generate
dc.generate:
	$(DOCKER_COMPOSE) run --rm generate

define make-generate-rules

$1.generate: bin/protoc-gen-go bin/protoc-gen-go-grpc bin/protoc-gen-grpc-gateway bin/protoc-gen-grpc-sarama bin/mockgen
	protoc \
		-I . \
		-I ./pkg/pb \
		-I $(dir $(shell (go list -f '{{ .Dir }}' github.com/justin0u0/protoc-gen-grpc-sarama/proto))) \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		--grpc-gateway_out=paths=source_relative:. \
		--grpc-sarama_out=paths=source_relative:. \
		./modules/$1/pb/*.proto

	go generate ./modules/$1/...

endef
$(foreach module,$(MODULES),$(eval $(call make-generate-rules,$(module))))

pkg.generate: bin/protoc-gen-go bin/protoc-gen-go-grpc bin/protoc-gen-grpc-gateway bin/protoc-gen-grpc-sarama bin/mockgen
	go generate ./pkg/...

generate: pkg.generate $(addsuffix .generate,$(MODULES))

bin/protoc-gen-go: go.mod
	go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go

bin/protoc-gen-go-grpc: go.mod
	go build -o $@ google.golang.org/grpc/cmd/protoc-gen-go-grpc

bin/protoc-gen-grpc-gateway: go.mod
	go build -o $@ github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway

bin/protoc-gen-grpc-sarama: go.mod
	go build -o $@ github.com/justin0u0/protoc-gen-grpc-sarama

bin/mockgen: go.mod
	go build -o $@ github.com/golang/mock/mockgen

####################################################################################################
### Rule for the `lint` command
###

define make-dc-lint-rules

.PHONY: dc.$1.lint
dc.$1.lint:
	$(DOCKER_COMPOSE) run --rm lint make $1.lint
endef
$(foreach module,$(MODULES),$(eval $(call make-dc-lint-rules,$(module))))

.PHONY: dc.pkg.lint
dc.pkg.lint:
	$(DOCKER_COMPOSE) run --rm lint make pkg.lint

.PHONY: dc.lint
dc.lint:
	$(DOCKER_COMPOSE) run --rm lint

define make-lint-rules

$1.lint:
	golangci-lint run ./modules/$1/...

endef
$(foreach module,$(MODULES),$(eval $(call make-lint-rules,$(module))))

pkg.lint:
	golangci-lint run ./pkg/...

lint: pkg.lint $(addsuffix .lint,$(MODULES))

####################################################################################################
### Rule for the `test` command
###

define make-dc-test-rules

.PHONY: dc.$1.test

# to test individual module, override the command defined in the docker-compose.yml file
dc.$1.test:
	$(DOCKER_COMPOSE) run --rm test make $1.test

endef
$(foreach module,$(MODULES),$(eval $(call make-dc-test-rules,$(module))))

.PHONY: dc.pkg.test
dc.pkg.test:
	$(DOCKER_COMPOSE) run --rm test make pkg.test

.PHONY: dc.test
dc.test:
	$(DOCKER_COMPOSE) run --rm test

define make-test-rules

$1.test:
	go test -v -race ./modules/$1/...

endef
$(foreach module,$(MODULES),$(eval $(call make-test-rules,$(module))))

pkg.test:
	go test -v -race ./pkg/...

test: pkg.test $(addsuffix .test,$(MODULES))

####################################################################################################
### Rule for the `build` command
###

.PHONY: dc.image
dc.image: dc.build
	$(DOCKER_COMPOSE) build --force-rm image

.PHONY: dc.build
dc.build:
	$(DOCKER_COMPOSE) run --rm build

build:
	mkdir -p ./bin/app
	go build -o ./bin/app/cmd ./cmd/main.go
