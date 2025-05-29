ifeq ($(OS),Windows_NT)
IS_WINDOWS := true
_DOCKER_BUILDX_CHECK_COMMAND := powershell -Command "(docker buildx inspect > $null 2>$null) -and ($LASTEXITCODE -eq 0)" || (docker buildx inspect >NUL 2>&1 && echo true || echo false)
else
IS_WINDOWS := false
_DOCKER_BUILDX_CHECK_COMMAND := docker buildx inspect >/dev/null 2>&1 && echo true || echo false
endif

DOCKER_BUILDX_PRESENT := $(shell $(_DOCKER_BUILDX_CHECK_COMMAND))

.PHONY: all docker-build docker-run help clean build-package publish

all: docker-build

help:
	@echo "Available targets:"
	@echo "  all            - Build the Docker image (default)."
	@echo "  docker-build   - Build the Docker image using Dockerfile."
	@echo "  docker-run     - Run the Docker container."
	@echo "  clean          - Remove build and distribution artifacts (dist/, build/, *.egg-info/)."
	@echo "  build-package  - Build Python package distribution archives (.whl, .tar.gz) using uv."
	@echo "  publish        - Upload Python package distributions to PyPI using uv. Requires UV_TOKEN."
	@echo "  help           - Show this help message."

DOCKERFILE := Dockerfile
DOCKER_IMAGE_NAME := gns3-api-util:latest

docker-build:
ifeq ($(strip $(DOCKER_BUILDX_PRESENT)),true)
	docker buildx build --load -t $(DOCKER_IMAGE_NAME) -f $(DOCKERFILE) .
else
	docker build --load -t $(DOCKER_IMAGE_NAME) -f $(DOCKERFILE) .
endif

docker-run:
ifeq ($(IS_WINDOWS),true)
	docker run -it --rm -p 8080:8080 -v "%USERPROFILE%\.gns3key:/root/.gns3key" $(DOCKER_IMAGE_NAME) $(ARGS) || (exit /b 0)
else
	docker run -it --rm -p 8080:8080 -v "$$HOME/.gns3key:/root/.gns3key" $(DOCKER_IMAGE_NAME) $(ARGS) || true
endif

clean:
	@echo "Cleaning up build and distribution artifacts..."
	$(RM) -r dist/ build/ *.egg-info/ || true

build-package: clean
	@echo "Building Python package distributions with uv..."
	uv venv
	uv build

publish: build-package
	@echo "Uploading Python package distributions to PyPI with uv..."
	uv publish
