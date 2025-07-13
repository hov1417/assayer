package assayer

import (
	"fmt"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/check"
	"github.com/hov1417/assayer/types"
	"strings"
)

func reportRepoResult(repo, verdictType, details string, verbose bool) error {
	var err error = nil
	if verbose && details != "" {
		_, err = fmt.Printf("%-60s %-40s %-40s\n", repo, verdictType, details)
	} else {
		_, err = fmt.Printf("%-60s %-40s\n", repo, verdictType)
	}
	return err
}

func reportResults(verdicts chan types.Response, args arguments.Arguments) error {
	for verdictRecord := range verdicts {
		if verdictRecord.Err != nil {
			return verdictRecord.Err
		}
		var err error = nil
		switch verdict := verdictRecord.Verdict.(type) {
		case types.Unmodified:
			err = reportRepoResult(verdict.Repository(), "Unmodified", "", args.Verbose)
		case check.Untracked:
			err = reportRepoResult(
				verdict.Repository(),
				"Untracked",
				fmt.Sprintf("Path \"%s\" is untracked", verdict.UntrackedItem()),
				args.Verbose)
		case check.Modified:
			err = reportRepoResult(
				verdict.Repository(),
				"Modified",
				fmt.Sprintf("File \"%s\" is %s", verdict.ModifiedItem(), types.Stringify(verdict.ModificationType())),
				args.Verbose)
		case check.LocalOnlyBranch:
			err = reportRepoResult(verdict.Repository(), "Local Only Branch", verdict.BranchName(), args.Verbose)
		case check.StashedChanges:
			err = reportRepoResult(verdict.Repository(),
				"Stashed Changes",
				fmt.Sprintf("on commit \"%s\"", firstLine(verdict.CommitUnderStash().Message)),
				args.Verbose)
		case check.RemoteAhead:
			err = reportRepoResult(verdict.Repository(),
				"Remote Mismatch",
				fmt.Sprintf("remote branch \"%s\" is ahead", verdict.LocalBranch()),
				args.Verbose)
		case check.RemoteBehind:
			err = reportRepoResult(verdict.Repository(),
				"Remote Mismatch",
				fmt.Sprintf("remote branch \"%s\" is behind", verdict.LocalBranch()),
				args.Verbose)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func ReportResultByCount(verdicts chan types.Response, arguments arguments.Arguments) error {
	untracked := 0
	modified := 0
	localOnlyBranch := 0
	stashedChanges := 0
	remoteAhead := 0
	remoteBehind := 0
	for verdictRecord := range verdicts {
		if verdictRecord.Err != nil {
			return verdictRecord.Err
		}
		switch verdictRecord.Verdict.(type) {
		case types.Unmodified:
		case check.Untracked:
			untracked += 1
		case check.Modified:
			modified += 1
		case check.LocalOnlyBranch:
			localOnlyBranch += 1
		case check.StashedChanges:
			stashedChanges += 1
		case check.RemoteAhead:
			remoteAhead += 1
		case check.RemoteBehind:
			remoteBehind += 1
		}
	}
	if arguments.Untracked {
		fmt.Printf("%-40s %d\n", "Repositories with Untracked files", untracked)
	}
	if arguments.Modified {
		fmt.Printf("%-40s %d\n", "Modified Repositories", modified)
	}
	if arguments.LocalOnlyBranch {
		fmt.Printf("%-40s %d\n", "Repositories With Local Only Branches", localOnlyBranch)
	}
	if arguments.StashedChanges {
		fmt.Printf("%-40s %d\n", "Repositories With Stashes", stashedChanges)
	}
	if arguments.RemoteAhead {
		fmt.Printf("%-40s %d\n", "Not Pulled Repositories", remoteAhead)
	}
	if arguments.RemoteBehind {
		fmt.Printf("%-40s %d\n", "Not Pushed Repositories", remoteBehind)
	}
	return nil
}

func firstLine(message string) string {
	newline := strings.IndexFunc(message, func(char rune) bool {
		return char == '\n' || char == '\r'
	})
	if newline == -1 {
		return message
	}
	return message[:newline]
}
