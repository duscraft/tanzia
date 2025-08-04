#!/bin/bash

set -euo pipefail

docker build . --file apps/front/Dockerfile --tag tanzia-front
docker login ghcr.io -u $GITHUB_ACTOR -p $PAT_TOKEN
docker tag tanzia-front ghcr.io/$REPO_OWNER/tanzia-front:latest
docker push ghcr.io/$REPO_OWNER/tanzia-front:latest
