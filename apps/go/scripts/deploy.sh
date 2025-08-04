#!/bin/bash

set -euo pipefail

docker build . --file apps/go/Dockerfile --tag tanzia
docker login ghcr.io -u $GITHUB_ACTOR -p $PAT_TOKEN
docker tag tanzia ghcr.io/$REPO_OWNER/tanzia:latest
docker push ghcr.io/$REPO_OWNER/tanzia:latest
