#!/bin/bash
set -e

rm -rf repo
mkdir repo
cd repo

git init -b main
git config user.name test
git config user.email test@test.com

BASE_DATE="2020-01-01T00:00:00Z"
DATE="$BASE_DATE"

increment_date() {
  DATE=$(TZ=UTC date -u -d "$DATE +1 day" +"%Y-%m-%dT%H:%M:%SZ")
}

export GIT_AUTHOR_NAME=test
export GIT_AUTHOR_EMAIL=test@test.com
export GIT_COMMITTER_NAME=test
export GIT_COMMITTER_EMAIL=test@test.com

mkdir -p docs

increment_date
echo "A" > docs/main.md
git add docs/main.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "A: main base commit"

increment_date
echo "B" >> docs/main.md
git add docs/main.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "B: main second commit"

increment_date
echo "C" >> docs/main.md
git add docs/main.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "C: main third commit"
