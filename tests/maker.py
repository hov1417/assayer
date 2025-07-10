#!/usr/bin/env python3

import argparse
import random
import subprocess
from pathlib import Path
from typing import List


def sh(cmd, cwd):
    subprocess.run(cmd, cwd=cwd, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def write(p: Path, text: str):
    p.write_text(text, encoding="utf-8")


def make_clean(repo: Path, other_args: List[str]):
    write(repo / "file.txt", "initial\n")
    sh(["git", "add", "."], repo)
    sh(["git", "commit", "-m", "initial commit"], repo)


def make_dirty(repo: Path, other_args: List[str]):
    write(repo / "file.txt", "changed but not staged\n")


def make_staged(repo: Path, other_args: List[str]):
    write(repo / "new.txt", "staged only\n")
    sh(["git", "add", "new.txt"], repo)


def make_stashed(repo: Path, other_args: List[str]):
    write(repo / "file.txt", "about to stash\n")
    sh(["git", "add", "file.txt"], repo)
    sh(["git", "stash", "--include-untracked", "-m", "work in progress"], repo)


def make_untracked(repo: Path, other_args: List[str]):
    write(repo / ("file" + str(random.randint(1, 100)) + ".txt"), "untracked file\n")


def clone(repo: Path, other_args: List[str]):
    remote_repo = other_args[0]
    sh(["git", "clone", remote_repo, repo / "repo"], repo)


SCENARIOS = {
    "clean": make_clean,
    "committed": make_clean,
    "dirty": make_dirty,
    "staged": make_staged,
    "stashed": make_stashed,
    "untracked": make_untracked,
    "clone": clone,
}


def main():
    parser = argparse.ArgumentParser(description="Create Git repo in a chosen state.")
    parser.add_argument("path", type=Path, help="target directory")
    parser.add_argument("state", choices=SCENARIOS, help="desired repo state")
    args, other_args = parser.parse_known_args()

    repo = args.path.resolve()
    repo.mkdir(parents=True, exist_ok=True)

    sh(["git", "init", "-q"], repo)

    SCENARIOS[args.state](repo, other_args)


if __name__ == "__main__":
    main()
