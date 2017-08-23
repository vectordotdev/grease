package main

import (
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"os"
)

/*
// The Accept header on GitHub HTTP API requests indicates
// the version of the API to use
const gitHubApiAccept = "application/vnd.github.v3+json"
const userAgentBase = "timberio/grease"
const urlBase = "https://api.github.com"

var userAgent = fmt.Sprintf("%s %s", userAgentBase, version)
*/

type gitHubRepo struct {
	Owner string
	Name  string
}

type gitHubRelease struct {
	TagName         *string `json:"tag_name"`
	TargetCommitish *string `json:"target_commitish,omitempty"`
	Name            *string `json:"name,omitempty"`
	Body            *string `json:"body,omitempty"`
	Draft           *bool   `json:"draft"`
	PreRelease      *bool   `json:"prerelease"`
}

func (repo *gitHubRepo) GetReleaseIdByTag(ctx context.Context, tag string, token string) (*int, error) {

	client := newGitHubAPIClient(ctx, token)
	release, _, err := client.Repositories.GetReleaseByTag(ctx, repo.Owner, repo.Name, tag)

	if err != nil {
		return nil, err
	}

	return release.ID, err
}

func (repo *gitHubRepo) CreateRelease(ctx context.Context, release *gitHubRelease, token string) (*int, error) {
	gRelease := &github.RepositoryRelease{
		TagName:         release.TagName,
		TargetCommitish: release.TargetCommitish,
		Name:            release.Name,
		Body:            release.Body,
		Draft:           release.Draft,
		Prerelease:      release.PreRelease,
	}

	client := newGitHubAPIClient(ctx, token)
	createdRelease, _, err := client.Repositories.CreateRelease(ctx, repo.Owner, repo.Name, gRelease)

	if err != nil {
		return nil, err
	}

	return createdRelease.ID, nil
}

func (repo *gitHubRepo) UpdateRelease(ctx context.Context, releaseId int, release *gitHubRelease, token string) (*int, error) {
	gRelease := &github.RepositoryRelease{
		Name:       release.Name,
		Body:       release.Body,
		Draft:      release.Draft,
		Prerelease: release.PreRelease,
	}

	client := newGitHubAPIClient(ctx, token)
	updatedRelease, _, err := client.Repositories.EditRelease(ctx, repo.Owner, repo.Name, releaseId, gRelease)

	if err != nil {
		return nil, err
	}

	return updatedRelease.ID, nil
}

func (repo *gitHubRepo) UploadReleaseAsset(ctx context.Context, releaseId int, file *os.File, filename string, token string) error {
	opts := &github.UploadOptions{
		Name: filename,
	}

	client := newGitHubAPIClient(ctx, token)
	_, _, err := client.Repositories.UploadReleaseAsset(ctx, repo.Owner, repo.Name, releaseId, opts, file)

	return err
}

func newGitHubAPIClient(ctx context.Context, token string) *github.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tokenClient := oauth2.NewClient(ctx, tokenSource)

	client := github.NewClient(tokenClient)

	return client
}
