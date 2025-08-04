#!/bin/bash

set -euo pipefail

sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no $SSH_USER@$SSH_IP -p $SSH_PORT 'sudo kubectl rollout restart deployment tanzia-app'
