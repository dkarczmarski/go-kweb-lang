#!/bin/bash

set -e

cd repo

BASE_DATE="2020-10-01T00:00:00Z"
DATE="$BASE_DATE"

increment_date() {
  DATE=$(TZ=UTC date -u -d "$DATE +1 day" +"%Y-%m-%dT%H:%M:%SZ")
}

increment_date
echo "X" >> file10.txt
git add file10.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (main) file10.txt"

increment_date
echo "X" >> file11.txt
git add file11.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (main) file11.txt"

git --no-pager log --graph --all --decorate --date=iso-strict --pretty=format:"%H %cd %s"
