#!/bin/bash

set -e

cd updater

git config --local user.name testuser
git config --local user.email testuser@foo.com

BASE_DATE=$(git log -1 --format=%cI)
DATE="$BASE_DATE"

increment_date() {
  DATE=$(TZ=UTC date -u -d "$DATE +1 day" +"%Y-%m-%dT%H:%M:%SZ")
}

export GIT_AUTHOR_NAME="testuser"
export GIT_AUTHOR_EMAIL="testuser@foo.com"
export GIT_COMMITTER_NAME="testuser"
export GIT_COMMITTER_EMAIL="testuser@foo.com"

git checkout -b feature

increment_date
echo "feature" > docs/feature.md
git add docs/feature.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "C: feature commit"

git checkout main

increment_date
echo "D" >> docs/main.md
git add docs/main.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "D: main third commit"

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff feature -m "M: merge feature"

git push origin main
