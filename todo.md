## question: git --pretty=format:%H %cd %s

```
--pretty=format:%H %cd %s
--pretty=format:"%H %cd %s"
```

you have to use "" in a command line but it works well when it is run from code.
should we modify code to work in a command line when you do "copy-pase"?

## issue 1

there is no merge point for commits made directly to the main branch.
gitseek should return file info with no MergePoint.

for example:
commit 2a4a506919b01acaffbd33fc09928ae217454b97 was made directly to the main branch and should have no MergePoint
but
commit 9ae1e9c13d69c5164c84ba10f1060d5f9d541214 was made at separate branch and then merged, 
so it should have some MergePoint.

### a sample solution

we can check if it is a direct commit as follows:

```
$ git --no-pager log main --pretty=format:"%H %cd %s" --date=iso-strict --first-parent | grep 2a4a506919b01acaffbd33fc09928ae217454b97
2a4a506919b01acaffbd33fc09928ae217454b97 2023-09-23T16:20:57-07:00 add link for kubelet and cloud-controller-manager (#40931)

$ git --no-pager log main --pretty=format:"%H %cd %s" --date=iso-strict --first-parent | grep 9ae1e9c13d69c5164c84ba10f1060d5f9d541214
(no result)
```

## issue 2

after the change in the file selection condition, 
files that have only a PR with no update can now appear on the list.
the issue is that a PR may appear on the list even after it's closed, 
because it's still treated as open. 
this happens because we fetch only open PRs, and there's no cleanup mechanism.

## issue 3

when we detect updates, it's better to use the date of the fork point 
rather than the commit date in the branch. 
