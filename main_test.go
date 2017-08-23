package main

import (
	"testing"
)

func TestSplitRepositoryName(test *testing.T) {
	expectedOwner := "timberio"
	expectedRepo := "grease"

	owner, repo, err := splitRepositoryName("timberio/grease")

	if err != nil {
		test.Fatalf("Did not expect to receive error: %v", err)
	}

	if owner != expectedOwner {
		test.Fatalf("Expected owner to be %s but got %s", expectedOwner, owner)
	}

	if repo != expectedRepo {
		test.Fatalf("Expected repo to be %s but got %s", expectedRepo, repo)
	}
}
