# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Downloads bundle ingredients : containerd as tar, kubelet, kubeadm, kubectl as Debian packages
#
# Usage:
# 1. Mount a host path as /ingredients
# 2. Run the image
#

ARG BASE_IMAGE=ubuntu:20.04
FROM $BASE_IMAGE as build

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends sudo

# Override to download other version
ENV CONTAINERD_VERSION=1.7.21
ENV KUBERNETES_VERSION=1.31.0-1.1
ENV ARCH=amd64

WORKDIR /bundle-builder
COPY download.sh .
RUN chmod a+x download.sh
WORKDIR /ingredients

ENTRYPOINT ["/bundle-builder/download.sh"]
