#!/bin/bash
set -e

cd repo

LAST_DATE=$(git log -1 --format=%cI)
NEXT_DATE=$(TZ=UTC date -u -d "$LAST_DATE +1 day" +"%Y-%m-%dT%H:%M:%SZ")

export GIT_AUTHOR_NAME=test
export GIT_AUTHOR_EMAIL=test@test.com
export GIT_COMMITTER_NAME=test
export GIT_COMMITTER_EMAIL=test@test.com

echo "C" >> docs/main.md
git add docs/main.md

GIT_AUTHOR_DATE=$NEXT_DATE GIT_COMMITTER_DATE=$NEXT_DATE \
git commit -m "C: main third commit"
