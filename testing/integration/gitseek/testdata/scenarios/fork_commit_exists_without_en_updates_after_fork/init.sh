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

mkdir -p content/en/docs
mkdir -p content/pl/docs

increment_date
echo "A" > content/en/docs/test.md
git add content/en/docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "A: add content/en/docs/test.md"

git checkout -b pl-branch

increment_date
echo "B" > content/pl/docs/test.md
git add content/pl/docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "B: add content/pl/docs/test.md on pl-branch"

git checkout main

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff --no-edit pl-branch

git --no-pager log --graph --all --decorate --date=iso-strict --pretty=format:"%H %cd %s"
