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

increment_date
touch init.txt
git add init.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "init commit"

increment_date
echo "X" >> file1.txt
git add file1.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (main) file1.txt"

increment_date
echo "X" >> file2.txt
git add file2.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (main) file2.txt"

increment_date
echo "X" >> file1.txt
git add file1.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (main) file1.txt 2"

git checkout -b branch1

increment_date
echo "X" >> file4.txt
git add file4.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (branch1) file4.txt"

git branch branch2
git branch branch3

git checkout branch2

increment_date
echo "X" >> file6.txt
git add file6.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (branch2) file6.txt"

git checkout branch3

increment_date
echo "X" >> file8.txt
git add file8.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (branch3) file8.txt"

git checkout branch1

increment_date
echo "X" >> file4.txt
git add file4.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (branch1) file4.txt 2"

increment_date
echo "X" >> file5.txt
git add file5.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (branch1) file5.txt"

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff --no-edit branch2

increment_date
echo "X" >> file2.txt
git add file2.txt
echo "X" >> file7.txt
git add file7.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (branch1) file2.txt file7.txt"

git checkout main

increment_date
echo "X" >> file3.txt
git add file3.txt
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git commit -m "commit (main) file3.txt"

git checkout main

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff --no-edit branch1

increment_date
GIT_AUTHOR_DATE=$DATE GIT_COMMITTER_DATE=$DATE git merge --no-ff --no-edit branch3

git --no-pager log --graph --all --decorate --date=iso-strict --pretty=format:"%H %cd %s"
