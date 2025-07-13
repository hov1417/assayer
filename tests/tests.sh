#!/usr/bin/env bats

setup_file() {
  mkdir "tests/repos" -p
}

teardown_file() {
  rm -rf tests/repos
}

@test "local commited repos" {
  ./tests/maker.py tests/repos/test1/repo1 clean
  ./tests/maker.py tests/repos/test1/repo2 clean
  ./tests/maker.py tests/repos/test1/repo3 clean
  expected="repo1                                                        Unmodified
repo2                                                        Unmodified
repo3                                                        Unmodified"
  result="$(go run . --untracked --unmodified --stashed tests/repos/test1 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "local commited repos with exclude" {
  ./tests/maker.py tests/repos/test2/repo1 clean
  ./tests/maker.py tests/repos/test2/repo2 clean
  ./tests/maker.py tests/repos/test2/repo3 clean
  ./tests/maker.py tests/repos/test2/repo4 clean
  expected="repo1                                                        Unmodified
repo2                                                        Unmodified
repo3                                                        Unmodified"
  result="$(go run . --exclude "*/repo4" --untracked --unmodified --stashed tests/repos/test2 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "local stashed repos with exclude" {
  ./tests/maker.py tests/repos/test3/repo1 clean
  ./tests/maker.py tests/repos/test3/repo1 stashed
  ./tests/maker.py tests/repos/test3/repo2 clean
  ./tests/maker.py tests/repos/test3/repo2 stashed
  ./tests/maker.py tests/repos/test3/repo3 clean
  ./tests/maker.py tests/repos/test3/repo3 stashed
  ./tests/maker.py tests/repos/test3/repo4 clean
  ./tests/maker.py tests/repos/test3/repo4 stashed
  expected='repo1                                                        Stashed Changes
repo2                                                        Stashed Changes
repo3                                                        Stashed Changes'
  result="$(go run . --exclude "*/repo4" --untracked --unmodified --stashed tests/repos/test3 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "remotes" {
  ./tests/maker.py tests/repos/test4/repo1 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test4/repo1/repo pop-commit
  ./tests/maker.py tests/repos/test4/repo2 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test4/repo2/repo committed
  ./tests/maker.py tests/repos/test4/repo3 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test4/repo3/repo pop-commit
  ./tests/maker.py tests/repos/test4/repo3/repo committed
  expected='repo1/repo                                                   Remote Ahead
repo2/repo                                                   Remote Behind
repo3/repo                                                   Remote Ahead'
  result="$(go run . --untracked --unmodified --stashed --ahead-branches --behind-branches --local-only-branches tests/repos/test4 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}

@test "deep" {
  ./tests/maker.py tests/repos/test5/repo1 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test5/repo1/repo pop-commit
  ./tests/maker.py tests/repos/test5/repo1/repo stashed
  ./tests/maker.py tests/repos/test5/repo1/repo untracked
  ./tests/maker.py tests/repos/test5/repo1/repo dirty
  ./tests/maker.py tests/repos/test5/repo1/repo staged
  ./tests/maker.py tests/repos/test5/repo2 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test5/repo2/repo committed
  ./tests/maker.py tests/repos/test5/repo3 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test5/repo3/repo pop-commit
  ./tests/maker.py tests/repos/test5/repo3/repo committed
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
  ./tests/maker.py tests/repos/test6/repo1 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test6/repo1/repo stashed
  ./tests/maker.py tests/repos/test6/repo1/repo/tests/repos clone git@github.com:hov1417/assayer.git
  expected='repo1/repo/tests/repos/repo                                  Unmodified
repo1/repo                                                   Untracked'
  result="$(go run . --all --nested tests/repos/test6 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}


@test "un nested" {
  ./tests/maker.py tests/repos/test7/repo1 clone git@github.com:hov1417/assayer.git
  ./tests/maker.py tests/repos/test7/repo1/repo stashed
  ./tests/maker.py tests/repos/test7/repo1/repo/tests/repos clone git@github.com:hov1417/assayer.git
  expected='repo1/repo                                                   Untracked'
  result="$(go run . --all tests/repos/test7 | sort)"
  echo "$result"
  [ "$result" = "$expected" ]
}
