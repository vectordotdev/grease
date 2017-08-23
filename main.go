package main

import (
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path"
	"strings"
)

var version string

type incorrectArgumentNumberError struct {
	expected int
	received int
}

type missingRequiredArgumentError struct {
	argument string
}

type badArgumentError struct {
	argument string
	reason   string
}

func main() {
	app := cli.NewApp()

	// Global flags

	debugFlag := cli.BoolFlag{
		Name:   "debug, d",
		Usage:  "prints out verbose statements about what grease is doing",
		EnvVar: "DEBUG",
	}

	dryRunFlag := cli.BoolFlag{
		Name:  "dry-run, n",
		Usage: "prevents changes from being made; best used with --debug to see what changes would be made, if any",
	}

	// Common, non-global flags

	assetsFlag := cli.StringFlag{
		Name:  "assets",
		Usage: "uploads the assets at the given path (glob patterns enabled)",
	}

	gitHubTokenFlag := cli.StringFlag{
		Name:   "github-token",
		Usage:  "used to authenticate the request with the GitHub API",
		EnvVar: "GITHUB_TOKEN",
	}

	draftFlag := cli.BoolFlag{
		Name:  "draft",
		Usage: "marks the release as a draft (unpublished)",
	}

	prereleaseFlag := cli.BoolFlag{
		Name:  "pre-release, pre",
		Usage: "marks the release as a pre-release",
	}

	nameFlag := cli.StringFlag{
		Name:  "name",
		Usage: "sets the name of the release, for example \"v0.4.0 - 2017-08-22\"",
	}

	notesFlag := cli.StringFlag{
		Name:  "notes",
		Usage: "sets the body of the release notes",
	}

	// Hidden flags

	// These flags are hidden from the user and are used to hold positional
	// arguments passed after flags

	globPatternFlag := cli.StringFlag{
		Name:   "glob-pattern",
		Usage:  "Glob pattern to find files with",
		Hidden: true,
	}

	repositoryFlag := cli.StringFlag{
		Name:   "repository",
		Usage:  "name of the GitHub repository to operate on (not including the owner portion)",
		Hidden: true,
	}

	ownerFlag := cli.StringFlag{
		Name:   "owner",
		Usage:  "owner of the GitHub repository",
		Hidden: true,
	}

	tagFlag := cli.StringFlag{
		Name:   "tag",
		Usage:  "tag to modify",
		Hidden: true,
	}

	targetCommittishFlag := cli.StringFlag{
		Name:   "target-commitish",
		Usage:  "a commit-ish identifier to create the tag from",
		Value:  "master",
		Hidden: true,
	}

	// Commands

	// createReleaseCommand

	createReleaseCommand := cli.Command{
		Name:      "create-release",
		Usage:     "creates a release on GitHub",
		ArgsUsage: "REPO TAG COMMITISH",
		Description: `
Creates a new GitHub release identified by TAG on the repository identified
by REPO using the COMMITTISH identifier.
`,
		Action: cmdCreateRelease,
		Before: beforeCreateRelease,
		Flags: []cli.Flag{
			repositoryFlag,
			ownerFlag,
			tagFlag,
			targetCommittishFlag,
			nameFlag,
			notesFlag,
			draftFlag,
			prereleaseFlag,
			assetsFlag,
			gitHubTokenFlag,
		},
	}

	// updateReleaseCommand

	updateReleaseCommand := cli.Command{
		Name:      "update-release",
		Usage:     "updates a release on GitHub",
		ArgsUsage: "REPO TAG",
		Description: `
Updates the GitHub release identified by TAG on the repository identified by
REPO based on the flags passed on the command line.
`,
		Action: cmdUpdateRelease,
		Before: beforeUpdateRelease,
		Flags: []cli.Flag{
			repositoryFlag,
			ownerFlag,
			tagFlag,
			nameFlag,
			notesFlag,
			draftFlag,
			prereleaseFlag,
			assetsFlag,
			gitHubTokenFlag,
		},
	}

	// uploadArtifactsCommand

	uploadArtifactsCommand := cli.Command{
		Name:      "upload-assets",
		Usage:     "uploads assets to an existing release on GitHub",
		ArgsUsage: "REPO TAG GLOB_PATTERN",
		Description: `
Takes all files found using the glob pattern at GLOB_PATTERN and uploads them as
assets for the GitHub release identified by TAG on the repository identified
by REPO.
`,
		Action: cmdUploadArtifacts,
		Before: beforeUploadArtifacts,
		Flags: []cli.Flag{
			globPatternFlag,
			repositoryFlag,
			ownerFlag,
			tagFlag,
			gitHubTokenFlag,
		},
	}

	// listFilesCommand

	listFilesCommand := cli.Command{
		Name:      "list-files",
		Usage:     "print out list of files found using the glob pattern",
		ArgsUsage: "GLOB_PATTERN",
		Description: `
Takes a GLOB_PATTERN and prints out a list of the matching files.

This command is designed for trouble-shooting the use of a glob pattern when
using them to specify assets to upload.
`,
		Action: cmdListFiles,
		Before: beforeListFiles,
		Flags: []cli.Flag{
			globPatternFlag,
		},
	}

	app.Usage = "creates and updates releases on GitHub with assets"
	app.Version = version

	app.Flags = []cli.Flag{
		debugFlag,
		dryRunFlag,
	}

	app.Commands = []cli.Command{
		createReleaseCommand,
		updateReleaseCommand,
		uploadArtifactsCommand,
		listFilesCommand,
	}

	app.Run(os.Args)
}

/*
Architectural pattern for the command interface

We use a package called `cli` for exposing the functionality of grease via
a command-line tool. The package defines the concept of an "application"; to
this application we add many "commands". The package is responsible for parsing
commands and arguments.

Each command has an action that it executes after parsing. The action is a function
of the form `func(*cli.Context) (error)`. We name these functions based on the
command name. The command "upload-assets", for example, uses the function
cmdUploadArtifacts as its function. Likewise, "create-release" uses the function
cmdCreateRelease.
*/

/*
before functions - func(*cli.Context)

The functions in this section are run prior to the command action but after
the context has been setup.

The intent of these before functions is to perform validation of command line
arguments and exit early if required parameters are not present or do not meet
constraints.
*/

func beforeCreateRelease(ctx *cli.Context) error {
	debug := ctx.GlobalBool("debug")
	// Expected positional arguments (3): REPO TAG COMMITTISH
	err := validatePositionalArgumentCount(ctx, 3)

	if err != nil {
		return err
	}

	arguments := ctx.Args()

	repo := arguments.Get(0)
	repoOwner, repoName, err := splitRepositoryName(repo)

	if err != nil {
		return err
	}

	err = ctx.Set("owner", repoOwner)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("GitHub repository owner is: %s\n", repoOwner)
	}

	err = ctx.Set("repository", repoName)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("GitHub repository name is: %s\n", repo)
	}

	tag := arguments.Get(1)
	err = ctx.Set("tag", tag)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Git tag is: %s\n", tag)
	}

	commitish := arguments.Get(2)
	err = ctx.Set("target-commitish", commitish)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Git commitish is: %s\n", commitish)
	}

	return nil
}

func beforeUpdateRelease(ctx *cli.Context) error {
	debug := ctx.GlobalBool("debug")
	// Expected positional arguments (2): REPO TAG
	err := validatePositionalArgumentCount(ctx, 2)

	if err != nil {
		return err
	}

	arguments := ctx.Args()

	repo := arguments.Get(0)
	repoOwner, repoName, err := splitRepositoryName(repo)

	if err != nil {
		return err
	}

	err = ctx.Set("owner", repoOwner)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("GitHub repository owner is: %s\n", repoOwner)
	}

	err = ctx.Set("repository", repoName)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("GitHub repository name is: %s\n", repo)
	}

	tag := arguments.Get(1)
	err = ctx.Set("tag", tag)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Git tag is: %s\n", tag)
	}

	return nil
}

func beforeUploadArtifacts(ctx *cli.Context) error {
	debug := ctx.GlobalBool("debug")
	// Expected positional arguments (3): REPO TAG GLOB_PATTERN
	err := validatePositionalArgumentCount(ctx, 3)

	if err != nil {
		return err
	}

	arguments := ctx.Args()

	repo := arguments.Get(0)
	repoOwner, repoName, err := splitRepositoryName(repo)

	if err != nil {
		return err
	}

	err = ctx.Set("owner", repoOwner)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("GitHub repository owner is: %s\n", repoOwner)
	}

	err = ctx.Set("repository", repoName)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("GitHub repository name is: %s\n", repo)
	}

	tag := arguments.Get(1)
	err = ctx.Set("tag", tag)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Git tag is: %s\n", tag)
	}

	globPattern := arguments.Get(2)
	err = ctx.Set("glob-pattern", globPattern)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Glob pattern is: %s\n", globPattern)
	}

	return nil
}

func beforeListFiles(ctx *cli.Context) error {
	debug := ctx.GlobalBool("debug")

	// Expected positional arguments (1): GLOB_PATTERN
	err := validatePositionalArgumentCount(ctx, 1)

	if err != nil {
		return err
	}

	arguments := ctx.Args()

	globPattern := arguments.Get(0)
	err = ctx.Set("glob-pattern", globPattern)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Glob pattern is: %s\n", globPattern)
	}

	return nil
}

func cmdCreateRelease(ctx *cli.Context) error {
	dry := ctx.GlobalBool("dry-run")
	debug := ctx.GlobalBool("debug")

	if debug {
		fmt.Println("Preparing to create release")
	}

	repoName := ctx.String("repository")
	repoOwner := ctx.String("owner")

	tagName := ctx.String("tag")
	targetCommitish := ctx.String("target-commitish")
	releaseName := ctx.String("name")
	releaseBody := ctx.String("notes")
	draft := ctx.Bool("draft")
	preRelease := ctx.Bool("pre-release")

	assetGlobPattern := ctx.String("assets")

	gitHubToken := ctx.String("github-token")

	repo := &gitHubRepo{
		Name:  repoName,
		Owner: repoOwner,
	}

	release := &gitHubRelease{
		TagName:         &tagName,
		TargetCommitish: &targetCommitish,
		Name:            &releaseName,
		Body:            &releaseBody,
		Draft:           &draft,
		PreRelease:      &preRelease,
	}

	assets, err := findFiles(assetGlobPattern)

	if err != nil {
		return err
	}

	if debug {
		fmt.Println("Will create release with the following settings...")
		fmt.Printf("Repo: https://github.com/%s/%s\n", repo.Owner, repo.Name)
		fmt.Printf("Using token: %s\n", gitHubToken)
		fmt.Printf("Tag: %s\n", release.TagName)
		fmt.Printf("Tag Commit: %s\n", release.TargetCommitish)
		fmt.Printf("Release Name/Title: %s\n", release.Name)
		fmt.Printf("Draft: %t\n", release.Draft)
		fmt.Printf("Pre-release: %t\n", release.PreRelease)
		fmt.Println("-----Begin Release Notes-----")
		fmt.Printf("%s\n", release.Body)
		fmt.Println("-----End Release Notes-----")

		if len(assets) == 0 {
			fmt.Println("No assets found to upload")
		} else {
			fmt.Println("The following assets will be uploaded:")
			for _, assetPath := range assets {
				filename := path.Base(assetPath)
				fmt.Printf("\t%s as %s\n", assetPath, filename)
			}
		}
	}

	if dry {
		fmt.Println("Dry run specified. Exiting.")
		return nil
	}

	netCtx := context.Background()

	releaseId, err := repo.CreateRelease(netCtx, release, gitHubToken)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Created release (id: %d)!\n", *releaseId)
		fmt.Println("Preparing to upload any assets")
	}

	for _, assetPath := range assets {
		filename := path.Base(assetPath)
		file, err := os.Open(assetPath)

		if err != nil {
			fmt.Printf("Failed to open %s. Skipping.\n", assetPath)
			continue
		}

		if debug {
			fmt.Printf("Uploading asset at %s as %s\n", assetPath, filename)
		}

		err = repo.UploadReleaseAsset(netCtx, *releaseId, file, filename, gitHubToken)

		if err != nil {
			fmt.Printf("Error while uploading asset at %s\n", filename)
			fmt.Println(err)
		}
	}

	return nil
}

func cmdUpdateRelease(ctx *cli.Context) error {
	dry := ctx.GlobalBool("dry-run")
	debug := ctx.GlobalBool("debug")

	if debug {
		fmt.Println("Preparing to update release")
	}

	repoName := ctx.String("repository")
	repoOwner := ctx.String("owner")

	tagName := ctx.String("tag")
	releaseName := ctx.String("name")
	releaseBody := ctx.String("notes")
	draft := ctx.Bool("draft")
	preRelease := ctx.Bool("pre-release")

	assetGlobPattern := ctx.String("assets")

	gitHubToken := ctx.String("github-token")

	repo := &gitHubRepo{
		Name:  repoName,
		Owner: repoOwner,
	}

	release := &gitHubRelease{
		TagName:    &tagName,
		Name:       &releaseName,
		Body:       &releaseBody,
		Draft:      &draft,
		PreRelease: &preRelease,
	}

	assets, err := findFiles(assetGlobPattern)

	if err != nil {
		return err
	}

	if debug {
		fmt.Println("Will update release with the following settings...")
		fmt.Printf("Repo: https://github.com/%s/%s\n", repo.Owner, repo.Name)
		fmt.Printf("Using token: %s\n", gitHubToken)
		fmt.Printf("Tag: %s\n", release.TagName)
		fmt.Printf("Release Name/Title: %s\n", release.Name)
		fmt.Printf("Draft: %t\n", release.Draft)
		fmt.Printf("Pre-release: %t\n", release.PreRelease)
		fmt.Println("-----Begin Release Notes-----")
		fmt.Printf("%s\n", release.Body)
		fmt.Println("-----End Release Notes-----")

		if len(assets) == 0 {
			fmt.Println("No assets found to upload")
		} else {
			fmt.Println("The following assets will be uploaded:")
			for _, assetPath := range assets {
				filename := path.Base(assetPath)
				fmt.Printf("\t%s as %s\n", assetPath, filename)
			}
		}
	}

	if dry {
		fmt.Println("Dry run specified. Exiting.")
		return nil
	}

	netCtx := context.Background()

	releaseId, err := repo.GetReleaseIdByTag(netCtx, *release.TagName, gitHubToken)

	if err != nil {
		return err
	}

	_, err = repo.UpdateRelease(netCtx, *releaseId, release, gitHubToken)

	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("Updated release (id: %d)!\n", *releaseId)
		fmt.Println("Preparing to upload any assets")
	}

	for _, assetPath := range assets {
		filename := path.Base(assetPath)
		file, err := os.Open(assetPath)

		if err != nil {
			fmt.Printf("Failed to open %s. Skipping.\n", assetPath)
			continue
		}

		if debug {
			fmt.Printf("Uploading asset at %s as %s\n", assetPath, filename)
		}

		err = repo.UploadReleaseAsset(netCtx, *releaseId, file, filename, gitHubToken)

		if err != nil {
			fmt.Printf("Error while uploading asset at %s\n", filename)
			fmt.Println(err)
		}
	}

	return nil
}

func cmdUploadArtifacts(ctx *cli.Context) error {
	dry := ctx.GlobalBool("dry-run")
	debug := ctx.GlobalBool("debug")

	if debug {
		fmt.Println("Preparing to upload assets")
	}

	repoName := ctx.String("repository")
	repoOwner := ctx.String("owner")
	tagName := ctx.String("tag")

	assetGlobPattern := ctx.String("glob-pattern")

	gitHubToken := ctx.String("github-token")

	repo := &gitHubRepo{
		Name:  repoName,
		Owner: repoOwner,
	}

	assets, err := findFiles(assetGlobPattern)

	if err != nil {
		return err
	}

	if debug {
		fmt.Println("Uploading assets with the following settings...")
		fmt.Printf("Repo: https://github.com/%s/%s\n", repo.Owner, repo.Name)
		fmt.Printf("Using token: %s\n", gitHubToken)
		fmt.Printf("Tag: %s\n", tagName)
		if len(assets) == 0 {
			fmt.Println("No assets found to upload")
		} else {
			fmt.Println("The following assets will be uploaded:")
			for _, assetPath := range assets {
				filename := path.Base(assetPath)
				fmt.Printf("\t%s as %s\n", assetPath, filename)
			}
		}
	}

	if dry {
		fmt.Println("Dry run specified. Exiting.")
		return nil
	}

	netCtx := context.Background()

	releaseId, err := repo.GetReleaseIdByTag(netCtx, tagName, gitHubToken)

	if err != nil {
		return err
	}

	for _, assetPath := range assets {
		filename := path.Base(assetPath)
		file, err := os.Open(assetPath)

		if err != nil {
			fmt.Printf("Failed to open %s. Skipping.\n", assetPath)
			continue
		}

		if debug {
			fmt.Printf("Uploading asset at %s as %s\n", assetPath, filename)
		}

		err = repo.UploadReleaseAsset(netCtx, *releaseId, file, filename, gitHubToken)

		if err != nil {
			fmt.Printf("Error while uploading asset at %s\n", filename)
			fmt.Println(err)
		}
	}

	return nil
}

func cmdListFiles(ctx *cli.Context) error {
	globPattern := ctx.String("glob-pattern")

	files, err := findFiles(globPattern)

	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No matches found")
	}

	for _, fileName := range files {
		fmt.Printf("File match found: %s\n", fileName)
	}

	return nil
}

func splitRepositoryName(name string) (owner string, repo string, e error) {
	endOwnerIndex := strings.Index(name, "/")

	if endOwnerIndex == -1 {
		e = &badArgumentError{argument: "REPO", reason: "expected to be of the form owner/repo but found no /"}
		return
	}

	if endOwnerIndex == 0 {
		e = &badArgumentError{argument: "REPO", reason: "expected to be of the form owner/repo but owner portion was blank"}
		return
	}

	repoStartIndex := endOwnerIndex + 1
	repoEndIndex := len(name)

	if repoStartIndex > repoEndIndex {
		e = &badArgumentError{argument: "REPO", reason: "expected to be of the form owner/repo but repo portion was blank"}
		return
	}

	owner = name[:endOwnerIndex]
	repo = name[repoStartIndex:repoEndIndex]

	return
}

func validatePositionalArgumentCount(ctx *cli.Context, expected int) error {
	received := ctx.NArg()

	if received != expected {
		return &incorrectArgumentNumberError{expected: expected, received: received}
	}

	return nil
}

func validateGitHubToken(gitHubToken string) error {
	if gitHubToken == "" {
		return &missingRequiredArgumentError{argument: "--github-token"}
	}

	return nil
}

func (e *missingRequiredArgumentError) Error() string {
	message := fmt.Sprintf("Required argument \"%s\" was not provided", e.argument)
	return message
}

func (e *incorrectArgumentNumberError) Error() string {
	message := fmt.Sprintf("Expected %d positional arguments but received %d", e.expected, e.received)
	return message
}

func (e *badArgumentError) Error() string {
	message := fmt.Sprintf("Bad argument %s: %s", e.argument, e.reason)
	return message
}

func (e *missingRequiredArgumentError) ExitCode() int {
	return 64
}

func (e *incorrectArgumentNumberError) ExitCode() int {
	return 64
}

func (e *badArgumentError) ExitCode() int {
	return 64
}

func (e *badGlobPatternError) ExitCode() int {
	return 64
}
