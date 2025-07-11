package assayer

import (
	"fmt"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/check"
	"github.com/hov1417/assayer/types"
	"strings"
)

func reportResults(verdicts chan types.Response) error {
	for verdictRecord := range verdicts {
		if verdictRecord.Err != nil {
			return verdictRecord.Err
		}
		switch verdict := verdictRecord.Verdict.(type) {
		case types.Unmodified:
			fmt.Printf("%-60s| Unmodified\n", verdict.Repository())
		case check.Untracked:
			fmt.Printf("%-60s| Path \"%s\" is untracked\n", verdict.Repository(), verdict.UntrackedItem())
		case check.Modified:
			fmt.Printf("%-60s| File \"%s\" is %s\n", verdict.Repository(), verdict.ModifiedItem(), types.Stringify(verdict.ModificationType()))
		case check.LocalOnlyBranch:
			fmt.Printf("%-60s| Local Only Branch \"%s\"\n", verdict.Repository(), verdict.BranchName())
		case check.StashedChanges:
			fmt.Printf("%-60s| Stashed Changes on commit \"%s\"\n", verdict.Repository(), firstLine(verdict.CommitUnderStash().Message))
		case check.RemoteAhead:
			fmt.Printf("%-60s| Remote Mismatch, remote branch \"%s\" is ahead\n", verdict.Repository(), verdict.LocalBranch())
		case check.RemoteBehind:
			fmt.Printf("%-60s| Remote Mismatch, remote branch \"%s\" is behind\n", verdict.Repository(), verdict.LocalBranch())
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
