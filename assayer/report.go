package assayer

import (
	"fmt"
	"strings"
)

func reportResults(verdicts chan HandleResponse) error {
	for verdictRecord := range verdicts {
		if verdictRecord.err != nil {
			return verdictRecord.err
		}
		switch verdict := verdictRecord.verdict.(type) {
		case Unmodified:
			fmt.Printf("%s: Unmodified\n", verdict.Repository())
		case Untracked:
			fmt.Printf("%s: Path \"%s\" is untracked\n", verdict.Repository(), verdict.untrackedItem)
		case Modified:
			fmt.Printf("%s: File \"%s\" is %s\n", verdict.Repository(), verdict.modifiedItem, Stringify(verdict.modificationType))
		case LocalOnlyBranch:
			fmt.Printf("%s: Local Only Branch \"%s\"\n", verdict.Repository(), verdict.branchName)
		case StashedChanges:
			fmt.Printf("%s: Stashed Changes on commit \"%s\"\n", verdict.Repository(), firstLine(verdict.commitUnderStash.Message))
		case RemoteAhead:
			fmt.Printf("%s: Remote Mismatch, remote branch \"%s\" is ahead\n", verdict.Repository(), verdict.localBranch)
		case RemoteBehind:
			fmt.Printf("%s: Remote Mismatch, remote branch \"%s\" is behind\n", verdict.Repository(), verdict.localBranch)
		}
	}
	return nil
}

func ReportResultByCount(verdicts chan HandleResponse, arguments Arguments) error {
	untracked := 0
	modified := 0
	localOnlyBranch := 0
	stashedChanges := 0
	remoteAhead := 0
	remoteBehind := 0
	for verdictRecord := range verdicts {
		if verdictRecord.err != nil {
			return verdictRecord.err
		}
		switch verdictRecord.verdict.(type) {
		case Unmodified:
		case Untracked:
			untracked += 1
		case Modified:
			modified += 1
		case LocalOnlyBranch:
			localOnlyBranch += 1
		case StashedChanges:
			stashedChanges += 1
		case RemoteAhead:
			remoteAhead += 1
		case RemoteBehind:
			remoteBehind += 1
		}
	}
	if arguments.Untracked {
		fmt.Printf("Repositories with Untracked files: %d\n", untracked)
	}
	if arguments.Modified {
		fmt.Printf("Modified Repositories: %d\n", modified)
	}
	if arguments.LocalOnlyBranch {
		fmt.Printf("Repositories With Local Only Branches: %d\n", localOnlyBranch)
	}
	if arguments.StashedChanges {
		fmt.Printf("Repositories With Stashes: %d\n", stashedChanges)
	}
	if arguments.RemoteAhead {
		fmt.Printf("Not Pulled Repositories: %d\n", remoteAhead)
	}
	if arguments.RemoteBehind {
		fmt.Printf("Not Pushed Repositories: %d\n", remoteBehind)
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
