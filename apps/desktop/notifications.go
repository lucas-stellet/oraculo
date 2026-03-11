package main

import "github.com/gen2brain/beeep"

func notifyApprovalPending(project string) {
	_ = beeep.Notify(
		"Oraculo — Approval Required",
		"An agent in "+project+" is waiting for your approval.",
		"",
	)
}

func notifyStoryCompleted(project, story string) {
	_ = beeep.Notify(
		"Oraculo — Story Completed",
		"Story '"+story+"' completed in "+project+".",
		"",
	)
}

func notifyEpicCompleted(project, epic string) {
	_ = beeep.Notify(
		"Oraculo — Epic Completed",
		"Epic '"+epic+"' completed in "+project+".",
		"",
	)
}
