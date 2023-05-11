#TARGET
TARGET=transbot
VERSION=0.1.1

TRANSBOT_PORT=8080
FRONTEND_PORT=8081

# Makefile for building docker image
DOCKER_TAG=$(TARGET):$(VERSION)

# build the image
DOCKERFILE_PATH=Dockerfile

build:
	@echo "Building: $(TARGET)"
	go build -o $(TARGET) .

docker:
	@echo "Building docker image with tag: $(DOCKER_TAG)"
	docker build -t $(DOCKER_TAG) -f $(DOCKERFILE_PATH) .

run:
	@echo "Running docker container with tag: $(DOCKER_TAG)"
	docker run -d -p 8080:$(TRANSBOT_PORT) -p 8081:$(FRONTEND_PORT)  $(DOCKER_TAG)

clean:
	@echo "Cleaning up images with tag: $(DOCKER_TAG)"
	docker image rm $(DOCKER_TAG)
