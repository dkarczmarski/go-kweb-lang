#!/bin/bash

set -e

cd repo

git config --local user.name testuser
git config --local user.email testuser@foo.com

DATE="2020-01-05T00:00:00Z"

export GIT_AUTHOR_NAME="testuser"
export GIT_AUTHOR_EMAIL="testuser@foo.com"
export GIT_COMMITTER_NAME="testuser"
export GIT_COMMITTER_EMAIL="testuser@foo.com"

echo "D" >> content/en/docs/test.md
git add content/en/docs/test.md
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "D: update content/en/docs/test.md again"

git --no-pager log --graph --all --decorate --date=iso-strict --pretty=format:"%H %cd %s"
