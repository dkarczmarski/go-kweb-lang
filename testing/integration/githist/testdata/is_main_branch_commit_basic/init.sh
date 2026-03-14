#!/bin/bash

set -e

rm -rf repo
mkdir -p repo
cd repo

git init -b main
git config --local user.name testuser
git config --local user.email testuser@foo.com

BASE_DATE="2020-01-01T00:00:00Z"
DATE="$BASE_DATE"

increment_date() {
  DATE=$(TZ=UTC date -u -d "$DATE +1 day" +"%Y-%m-%dT%H:%M:%SZ")
}

export GIT_AUTHOR_NAME="testuser"
export GIT_AUTHOR_EMAIL="testuser@foo.com"
export GIT_COMMITTER_NAME="testuser"
export GIT_COMMITTER_EMAIL="testuser@foo.com"

mkdir -p docs

increment_date
echo "A" > docs/main.md
git add docs/main.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "A: main base commit"

increment_date
echo "B" >> docs/main.md
git add docs/main.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "B: main second commit"
