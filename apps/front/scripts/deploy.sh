#!/bin/bash

set -euo pipefail

docker build -t tanzia-front -f apps/front/Dockerfile .
docker login ghcr.io -u $GITHUB_ACTOR -p $PAT_TOKEN
docker tag tanzia-front ghcr.io/$REPO_OWNER/tanzia-front:latest
docker push ghcr.io/$REPO_OWNER/tanzia-front:latest
