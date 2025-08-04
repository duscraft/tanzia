!#/bin/bash

set -e

docker build . --file Dockerfile --tag tanzia-front
docker login ghcr.io -u $GITHUB_ACTOR -p $PAT_TOKEN
docker tag tanzia-front ghcr.io/$REPO_OWNER/tanzia-front:latest
docker push ghcr.io/$REPO_OWNER/tanzia-front:latest
