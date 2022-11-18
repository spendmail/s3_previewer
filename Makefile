BIN := "./bin/previewer"
DOCKER_IMG="previewer:develop"
DOCKER_CONTAINER="previewer"
CONFIG_FILE_NAME="previewer.docker"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# Build application locally (without docker).
#build:
#	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/previewer

build:
	CGO_ENABLED=0 go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/previewer

# Launch application locally (without docker).
launch: build
	$(BIN) -config ./configs/previewer.toml

# Build docker image (without docker-compose).
build-img:
	docker build \
		--build-arg=CONFIG_FILE_NAME="$(CONFIG_FILE_NAME)" \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

# Launch application in docker container (without docker-compose).
run-img: build-img
	docker run -d --rm -p 8888:8888 --name $(DOCKER_CONTAINER) $(DOCKER_IMG)

# Stop application in docker container (without docker-compose).
stop-img:
	docker stop $(DOCKER_CONTAINER)

# Launch application in docker container using docker-compose.
run:
	LDFLAGS="$(LDFLAGS)" \
	CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) \
	docker-compose -f deployments/docker-compose.yaml up -d

# Stop application in docker container using docker-compose.
stop:
	LDFLAGS="$(LDFLAGS)" \
	CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) \
	docker-compose -f deployments/docker-compose.yaml down

# Launch integration testing using docker-compose.
test:
	set -e ;\
	LDFLAGS="$(LDFLAGS)" CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) docker-compose -f deployments/docker-compose.test.yaml up --build -d ;\
	test_status_code=0 ;\
	LDFLAGS="$(LDFLAGS)" CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) docker-compose -f deployments/docker-compose.test.yaml run integration_tests go test -v || test_status_code=$$? ;\
	LDFLAGS="$(LDFLAGS)" CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) docker-compose -f deployments/docker-compose.test.yaml down ;\
	exit $$test_status_code ;

.PHONY: build run build-img run-img test
