# Note: You can use any Debian/Ubuntu based image you want. 
FROM mcr.microsoft.com/devcontainers/base:bullseye

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

RUN export VERSION=4.18.3 \
    && curl -o migrate.deb -L https://github.com/golang-migrate/migrate/releases/download/v${VERSION}/migrate.linux-arm64.deb \
    && apt install ./migrate.deb \
    && rm migrate.deb

RUN export VERSION=1.86.1 \
    && curl -o tbls.deb -L https://github.com/k1LoW/tbls/releases/download/v${VERSION}/tbls_${VERSION}-1_arm64.deb \
    && apt install ./tbls.deb \
    && rm tbls.deb
