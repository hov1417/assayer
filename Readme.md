# Assayer

## Overview

Assayer is a CLI tool for checking Git repositories for uncompleted work. It can identify repositories with untracked files, modified files, stashed changes, branches that are ahead or behind the remote, and local-only branches.

## Installation

To install Assayer, you need to have Go installed. Then you can clone the repository and build the tool:

```sh
go install github.com/hov1417/assayer@latest
```

## Usage

Assayer can be run from the command line. Here is the basic usage:

```sh
assayer [options] [root-path-to-traverse]
```

### Options

- `--all, -a`: Check all in repositories.
- `--unmodified, -u`: Show repositories where nothing is changed.
- `--modified, -m`: Check if the worktree is changed.
- `--untracked, -t`: Check if there are untracked files.
- `--stashed, -s`: Check if there are stashed changes.
- `--behind-branches, -b`: Check if there are branches that are behind the remote.
- `--ahead-branches, -A`: Check if there are branches that are ahead of the remote.
- `--local-only-branches, -l`: Check if there are local-only branches.
- `--nested, -n`: Check repositories in repositories.
- `--count, -c`: Check repositories and report number of types.
- `--exclude, -e`: Exclude repositories, using [glob](https://github.com/gobwas/glob) patterns.
- `--reporter, -r`: Reporter's template using go's template syntax.

## Examples

Check all repositories in the current directory for any uncompleted work:

```sh
assayer --all
```

Check a specific directory for modified files and untracked files:

```sh
assayer --modified --untracked /path/to/check
```
or
```sh
assayer -mu /path/to/check
```

Check nested repositories:

```sh
assayer --nested
```

Using reporters:

```sh
JSON_TEMPLATE='{"unmodified":{{.unmodified}}, "untracked":{{.untracked}}, "modified":{{.modified}}, "localOnlyBranch":{{.localOnlyBranch}}, "stashedChanges":{{.stashedChanges}}, "remoteAhead":{{.remoteAhead}}, "remoteBehind":{{.remoteBehind}}}'
assayer -d -a -r $JSON_TEMPLATE /path/to/check
```

Reporters can be useful for shell prompts such as starship:

```sh
COLORED_TEMPLATE="{{if .untracked}}\e[38;5;88m{{.untracked}}{{end}}\
{{if .modified}}\e[0;91m{{.modified}}{{end}}\
{{if .localOnlyBranch}}\e[0;93m{{.localOnlyBranch}}{{end}}\
{{if .stashedChanges}}\e[0;92m{{.stashedChanges}}{{end}}\
{{if .remoteBehind}}\e[0;94m{{.remoteBehind}}{{end}}\
{{if .remoteAhead}}\e[0;35m{{.remoteAhead}}{{end}} "
assayer -d -a -r $COLORED_TEMPLATE /path/to/check
```
and then use something like this in your starship configurations
```toml
[custom.projects-status]
command = 'printf $(./path/to/cached/script)'
when = '[ "$PWD" = "$HOME" ]'
shell = ["bash", "--noprofile", "--norc"]
```


## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.