# Codespaces / Dev Container image for system-hub.
#
# Go itself is installed by the devcontainer "go" feature (see
# devcontainer.json) so the version stays pinned to go.mod. This image adds the
# system libraries Fyne needs to build — CGO plus OpenGL/X11 headers, the same
# set the CI release job installs — and a virtual X server (xvfb) so the GUI
# can be exercised headlessly in the cloud.
FROM mcr.microsoft.com/devcontainers/base:bookworm

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        gcc pkg-config \
        libgl1-mesa-dev libxrandr-dev libxcursor-dev libxi-dev \
        libxinerama-dev libxxf86vm-dev libxkbcommon-dev \
        xvfb xauth \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
