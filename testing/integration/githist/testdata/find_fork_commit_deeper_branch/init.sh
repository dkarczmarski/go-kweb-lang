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
echo "A" > docs/test.md
git add docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "A: main base commit"

increment_date
echo "B" >> docs/test.md
git add docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "B: main second commit"

git checkout -b feature

increment_date
echo "C" >> docs/test.md
git add docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "C: feature first commit"

increment_date
echo "D" >> docs/test.md
git add docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "D: feature second commit"

git checkout main

increment_date
echo "E" >> docs/test.md
git add docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "E: main third commit"
