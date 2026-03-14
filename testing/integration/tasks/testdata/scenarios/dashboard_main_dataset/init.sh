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
echo "EN unchanged" > content/en/docs/unchanged.md
echo "EN updated" > content/en/docs/updated.md
echo "EN deleted" > content/en/docs/deleted.md
echo "EN merged" > content/en/docs/merged.md
git add content/en/docs/unchanged.md content/en/docs/updated.md content/en/docs/deleted.md content/en/docs/merged.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "A: add initial EN docs"

increment_date
echo "PL unchanged" > content/pl/docs/unchanged.md
echo "PL updated" > content/pl/docs/updated.md
echo "PL deleted" > content/pl/docs/deleted.md
git add content/pl/docs/unchanged.md content/pl/docs/updated.md content/pl/docs/deleted.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "B: add initial PL docs on main"

git checkout -b pl-branch

increment_date
echo "PL merged" > content/pl/docs/merged.md
git add content/pl/docs/merged.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "C: add merged PL doc on pl-branch"

git checkout main

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff --no-edit pl-branch

increment_date
echo "EN updated after PL" >> content/en/docs/updated.md
git add content/en/docs/updated.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "D: update updated.md on main"

increment_date
git rm content/en/docs/deleted.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "E: delete deleted.md on main"

git checkout -b en-branch

increment_date
echo "EN merged branch update" >> content/en/docs/merged.md
git add content/en/docs/merged.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "F: update merged.md on en-branch"

git checkout main

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff --no-edit en-branch

git --no-pager log --graph --all --decorate --date=iso-strict --pretty=format:"%H %cd %s"
