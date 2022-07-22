
.DEFAULT_GOAL := docker-build-full

linux:
	@echo ">>> Make: Setting GO target OS to 'linux'"
	GOOS=linux
    export GOOS

darwin:
	@echo ">>> Make: Setting GO target OS to 'darwin'"
	GOOS=darwin
    export GOOS

windows:
	@echo ">>> Make: Setting GO target OS to 'windows'"
	GOOS=windows
    export GOOS

dep:
	@echo ">>> Make: Updating dependencies"
	dep ensure

build:
	@echo ">>> Make: Building all modules"
	$(MAKE) -C k8s-health
	$(MAKE) -C init
	$(MAKE) -C lock

docker-build:
	@echo ">>> Make: Building all docker images"
	$(MAKE) -C k8s-health docker-build
	$(MAKE) -C init docker-build
	$(MAKE) -C lock docker-build

docker-push:
	@echo ">>> Make: Pushing all docker images"
	$(MAKE) -C k8s-health docker-push
	$(MAKE) -C init docker-push
	$(MAKE) -C lock docker-push

docker-push-latest:
	@echo ">>> Make: Pushing all docker images with latest tag"
	$(MAKE) -C k8s-health docker-push-latest
	$(MAKE) -C init docker-push-latest
	$(MAKE) -C lock docker-push-latest


# App version for docker build
ver = version.info
include $(ver)
export $(shell sed 's/=.*//' $(ver))

# App info for docker build
app = app.info
include $(app)
export $(shell sed 's/=.*//' $(app))

docker-build-full:
	@echo ">>> Make: build docker images from docker builder"
	docker build -t $(DOCKER_USER)/$(DOCKER_NAME):$(DOCKER_NAME_INIT)-$(APP_VERSION) --target=init .
	docker build -t $(DOCKER_USER)/$(DOCKER_NAME):$(DOCKER_NAME_HEALTH)-$(APP_VERSION) --target=k8s-health .
	docker build -t $(DOCKER_USER)/$(DOCKER_NAME):$(DOCKER_NAME_LOCK)-$(APP_VERSION) --target=lock .

docker-push: docker-build-full
	@echo ">>> Make: Pushing all docker images"
	docker push $(DOCKER_USER)/$(DOCKER_NAME):$(DOCKER_NAME_INIT)-$(APP_VERSION) 
	docker push $(DOCKER_USER)/$(DOCKER_NAME):$(DOCKER_NAME_HEALTH)-$(APP_VERSION)
	docker push $(DOCKER_USER)/$(DOCKER_NAME):$(DOCKER_NAME_LOCK)-$(APP_VERSION)
