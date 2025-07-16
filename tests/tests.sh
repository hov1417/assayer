#!/usr/bin/env bats

make_clean() {
  (
    local repo="$1"

    mkdir -p "$repo"
    cd "$repo" || exit
    git init
    echo "commit" > "file.txt"
    git add .
    git commit -m "some commit"
  )
}

make_commit() {
  (
    local repo="$1"

    mkdir -p "$repo"
    cd "$repo" || exit
    echo "commit" > "file.txt"
    git add .
    git commit -m "some commit"
  )
}

make_dirty() {
  (
    local repo="$1"

    echo "changed but not staged" > "$repo/file.txt"
  )
}

make_staged() {
  (
    local repo="$1"

    mkdir -p "$repo"
    cd "$repo" || exit
    echo "staged only" > "new.txt"
    git add new.txt
  )
}

make_stashed() {
  (
    local repo="$1"

    mkdir -p "$repo"
    cd "$repo" || exit
    echo "about to stash" > "file.txt"
    git add file.txt
    git stash --include-untracked -m "work in progress"
  )
}

make_untracked() {
  (
    local repo="$1"

    local random_num=$((RANDOM % 100 + 1))
    echo "untracked file" > "$repo/file${random_num}.txt"
  )
}

make_branch() {
  (
    local repo="$1"

    mkdir -p "$repo"
    cd "$repo" || exit
    git branch new-branch
  )
}

clone() {
  (
    local repo="$1"
    local remote_repo="$2"

    mkdir -p "$repo"
    cd "$repo" || exit
    git clone "$remote_repo" "repo"
  )
}

pop_commit() {
  (
    local repo="$1"

    mkdir -p "$repo"
    cd "$repo" || exit
    git reset HEAD^ --hard
  )
}

setup_file() {
  rm -rf tests/repos
  mkdir "tests/repos" -p
}

#teardown_file() {
#  rm -rf tests/repos
#}

@test "local commited repos" {
  make_clean tests/repos/test1/repo1
  make_clean tests/repos/test1/repo2
  make_clean tests/repos/test1/repo3
  expected="repo1                                                        Unmodified
repo2                                                        Unmodified
repo3                                                        Unmodified"
  result="$(go run . --untracked --unmodified --stashed tests/repos/test1 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "local commited repos with exclude" {
  make_clean tests/repos/test2/repo1
  make_clean tests/repos/test2/repo2
  make_clean tests/repos/test2/repo3
  make_clean tests/repos/test2/repo4
  expected="repo1                                                        Unmodified
repo2                                                        Unmodified
repo3                                                        Unmodified"
  result="$(go run . --exclude "*/repo4" --untracked --unmodified --stashed tests/repos/test2 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "local stashed repos with exclude" {
  make_clean tests/repos/test3/repo1
  make_stashed tests/repos/test3/repo1
  make_clean tests/repos/test3/repo2
  make_stashed tests/repos/test3/repo2
  make_clean tests/repos/test3/repo3
  make_stashed tests/repos/test3/repo3
  make_clean tests/repos/test3/repo4
  make_stashed tests/repos/test3/repo4
  expected='repo1                                                        Stashed Changes
repo2                                                        Stashed Changes
repo3                                                        Stashed Changes'
  result="$(go run . --exclude "*/repo4" --untracked --unmodified --stashed tests/repos/test3 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "remotes" {
  clone tests/repos/test4/repo1 git@github.com:hov1417/assayer.git
  pop_commit tests/repos/test4/repo1/repo
  clone tests/repos/test4/repo2 git@github.com:hov1417/assayer.git
  make_commit tests/repos/test4/repo2/repo
  clone tests/repos/test4/repo3 git@github.com:hov1417/assayer.git
  pop_commit tests/repos/test4/repo3/repo
  make_commit tests/repos/test4/repo3/repo
  expected='repo1/repo                                                   Remote Ahead
repo2/repo                                                   Remote Behind
repo3/repo                                                   Remote Ahead'
  result="$(go run . --untracked --unmodified --stashed --ahead-branches --behind-branches --local-only-branches tests/repos/test4 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "deep" {
  clone tests/repos/test5/repo1 git@github.com:hov1417/assayer.git
  pop_commit tests/repos/test5/repo1/repo
  make_stashed tests/repos/test5/repo1/repo
  make_untracked tests/repos/test5/repo1/repo
  make_dirty tests/repos/test5/repo1/repo
  make_staged tests/repos/test5/repo1/repo
  clone tests/repos/test5/repo2 git@github.com:hov1417/assayer.git
  make_commit tests/repos/test5/repo2/repo
  clone tests/repos/test5/repo3 git@github.com:hov1417/assayer.git
  pop_commit tests/repos/test5/repo3/repo
  make_commit tests/repos/test5/repo3/repo
  expected='repo1/repo                                                   Modified
repo1/repo                                                   Remote Ahead
repo1/repo                                                   Stashed Changes
repo1/repo                                                   Untracked
repo2/repo                                                   Remote Behind
repo3/repo                                                   Remote Ahead'
  result="$(go run . --deep --all tests/repos/test5 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "nested" {
  clone tests/repos/test6/repo1 git@github.com:hov1417/assayer.git
  make_stashed tests/repos/test6/repo1/repo
  clone tests/repos/test6/repo1/repo/tests/repos git@github.com:hov1417/assayer.git
  expected='repo1/repo                                                   Stashed Changes
repo1/repo/tests/repos/repo                                  Unmodified'
  result="$(go run . --all --nested --deep tests/repos/test6 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}


@test "unnested" {
  clone tests/repos/test7/repo1 git@github.com:hov1417/assayer.git
  make_stashed tests/repos/test7/repo1/repo
  clone tests/repos/test7/repo1/repo/tests/repos git@github.com:hov1417/assayer.git
  expected='repo1/repo                                                   Stashed Changes'
  result="$(go run . --all tests/repos/test7 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "untracked" {
  clone tests/repos/test8/repo1 git@github.com:hov1417/assayer.git
  clone tests/repos/test8/repo2 git@github.com:hov1417/assayer.git
  clone tests/repos/test8/repo3 git@github.com:hov1417/assayer.git
  make_untracked tests/repos/test8/repo3/repo
  expected='repo3/repo                                                   Untracked'
  result="$(go run . --untracked tests/repos/test8 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "reporter" {
  clone tests/repos/test9/repo1 git@github.com:hov1417/assayer.git
  clone tests/repos/test9/repo2 git@github.com:hov1417/assayer.git
  clone tests/repos/test9/repo3 git@github.com:hov1417/assayer.git
  make_untracked tests/repos/test9/repo3/repo
  clone tests/repos/test9/repo4 git@github.com:hov1417/assayer.git
  pop_commit tests/repos/test9/repo4/repo
  make_stashed tests/repos/test9/repo4/repo
  make_untracked tests/repos/test9/repo4/repo
  make_dirty tests/repos/test9/repo4/repo
  make_staged tests/repos/test9/repo4/repo
  clone tests/repos/test9/repo5 git@github.com:hov1417/assayer.git
  make_commit tests/repos/test9/repo5/repo
  clone tests/repos/test9/repo6 git@github.com:hov1417/assayer.git
  pop_commit tests/repos/test9/repo6/repo
  make_commit tests/repos/test9/repo6/repo
  make_branch tests/repos/test9/repo6/repo
  expected='unmodified:2,untracked:2,modified:1,localOnlyBranch:1,stashedChanges:1,remoteAhead:2,remoteBehind:1'
  template="unmodified:{{.unmodified}},untracked:{{.untracked}},modified:{{.modified}},localOnlyBranch:{{.localOnlyBranch}},stashedChanges:{{.stashedChanges}},remoteAhead:{{.remoteAhead}},remoteBehind:{{.remoteBehind}}"
  result="$(go run . -d -a -r $template tests/repos/test9 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "gitignore" {
  make_clean tests/repos/test10/repo1
  echo "ignored" > tests/repos/test10/repo1/.gitignore
  make_clean tests/repos/test10/repo1
  mkdir tests/repos/test10/repo1/ignored
  touch tests/repos/test10/repo1/ignored/file1.txt
  touch tests/repos/test10/repo1/ignored/file2.txt
  expected='repo1                                                        Local Only Branch'
  result="$(go run . -d -a tests/repos/test10 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}


@test "gitignore2" {
  make_clean tests/repos/test11/repo1
  echo "ignored/" > tests/repos/test11/repo1/.gitignore
  make_clean tests/repos/test11/repo1
  mkdir tests/repos/test11/repo1/ignored
  touch tests/repos/test11/repo1/ignored/file1.txt
  touch tests/repos/test11/repo1/ignored/file2.txt
  expected='repo1                                                        Local Only Branch'
  result="$(go run . -d -a tests/repos/test11 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}


@test "gitignore3" {
  make_clean tests/repos/test12/repo1
  echo "/ignored/" > tests/repos/test12/repo1/.gitignore
  make_clean tests/repos/test12/repo1
  mkdir tests/repos/test12/repo1/ignored
  touch tests/repos/test12/repo1/ignored/file1.txt
  touch tests/repos/test12/repo1/ignored/file2.txt
  expected='repo1                                                        Local Only Branch'
  result="$(go run . -d -a tests/repos/test12 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}


@test "ignored nested repo" {
  make_clean tests/repos/test13/repo1
  echo "repo" > tests/repos/test13/repo1/.gitignore
  make_clean tests/repos/test13/repo1
  clone tests/repos/test13/repo1 git@github.com:hov1417/assayer.git
  expected='repo1                                                        Local Only Branch'
  result="$(go run . -d -a tests/repos/test13 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "local exclude" {
  make_clean tests/repos/test14/repo1
  echo "/ignored/" > tests/repos/test14/repo1/.git/info/exclude
  mkdir tests/repos/test14/repo1/ignored
  touch tests/repos/test14/repo1/ignored/file1.txt
  touch tests/repos/test14/repo1/ignored/file2.txt
  expected='repo1                                                        Local Only Branch'
  result="$(go run . -d -a tests/repos/test14 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "ignored nested dirty repo" {
  make_clean tests/repos/test15/repo1
  echo "repo" > tests/repos/test15/repo1/.gitignore
  make_clean tests/repos/test15/repo1
  clone tests/repos/test15/repo1 git@github.com:hov1417/assayer.git
  make_dirty tests/repos/test15/repo1/repo
  expected='repo1                                                        Local Only Branch'
  result="$(go run . -d -a tests/repos/test15 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}
