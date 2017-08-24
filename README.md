# Grease

[![GitHub release](https://img.shields.io/github/release/timberio/grease.svg)](https://github.com/timberio/grease/releases/latest) [![license](https://img.shields.io/github/license/timberio/grease.svg)](https://github.com/timberio/grease/blob/master/LICENSE) [![Github All Releases](https://img.shields.io/github/downloads/timberio/grease/total.svg)](https://github.com/timberio/grease/releases) [![CircleCI](https://img.shields.io/circleci/project/github/timberio/grease.svg)](https://circleci.com/gh/timberio/grease/tree/master)

Grease is a utility from Timber.io for creating and maintaining Github
releases. It is designed to be used in automated settings, like a CI
server.

## Installing

Pre-built binaries are available from the [repository's releases
page](https://github.com/timberio/grease/releases). Select the appropriate
package for your operating system. The binary is located under the `bin` folder
once you unpack the archive.

## Usage

In order to use Grease, you will need a GitHub personal access token. You can
create one from your [token settings](https://github.com/settings/tokens) page
on your profile. The token should be granted access to the general `repo` scope.

You can pass your GitHub access token to grease using the `--github-token` flag
after a sub-command or by setting the `GITHUB_TOKEN` environment variable.

The actual actions you take are performed using one of the sub-commands:

  * `create-release`
  * `update-release`
  * `upload-assets`
  * `list-files`

There are two global flags that can be passed directly after the `grease`
command and before any sub-command:

  * `--dry-run`, `-n` - will prepare any changes without actually applying them
  via the GitHub API.
  * `--debug`, `-d` - turns on verbose output.

The single-letter versions of the flags _cannot_ be combined into a
single parameter (like `-dn`) and must be passed separate (like `-d -n`).

### Creating a Release

You can create a release using the `create-release` sub-command which takes
three positional arguments: the repository name, the tag name, and the
[commit(ish)](https://git-scm.com/docs/gitglossary#gitglossary-aiddefcommit-ishacommit-ishalsocommittish)
the tag should be created from.

```shell
grease create-release timberio/grease v1.0.0 566e0e80a473581d05bcdf8a0cddad97ea2fc6c2
```

In addition to these, you can modify the release information using additional
flags. These flags should be listed between `create-release` and the positional
arguments:

  * `--name` - this flag is something that identifies the release to the user.
  Usually it's the name of the tag, but you might also include a date in the
  name (like "v0.4.0 - 2017-08-23") or a code-name if you're really cool
  (like "Sleeping Hyena").
  * `--notes` - Give this flag some text! The notes appear as the body of the
  release, and it's what users will see. It's a good idea to include
  information here about what has changed since the last release. You should
  take care to properly quote the value you pass to notes; if the notes aren't
  properly quoted, the shell might pass it to Grease as many additional
  parameters. You can pass the `--debug` and `--dry` global flags to make
  sure that the notes are picked up correctly.
  * `--draft` - this flag shouldn't be followed by a value. If it is present,
  the release will be set to "Draft" mode to be edited and released later on.
  * `--pre-release`, `--pre` - this flag shouldn't be followed by a value. If it
  is present, the release will be marked as a "Pre Release". You'll want to use
  this if you're releasing betas, nightlies, previews, or anything else that isn't
  considered part of the stable release set.
  * `--assets` - this flag takes a file glob pattern (like `"dist/*"`) for a
  value. Grease will try to upload any files matching the glob pattern as
  assets for your release. If you distribute pre-compiled binaries with your
  releases, this is the flag you want to use!

The final flag you need to know about is `--github-token`. A GitHub personal
access token is needed to create a release. More details are at the top of the
Usage section.

### Updating a Release

If you have already created a release (or pushed a git tag), you can update it
using Grease with the `update-release` sub-command which takes two positional
arguments: the repository name and the tag name.

```shell
grease update-release timberio/grease v1.1.0
```

This alone won't do much. To take full advantage of the sub-command, you will need
to pass additional flags to modify the release. The available flags are the same
as those available for `create-release` (detailed above) and should be specified
between `update-release` and the positional arguments. Please make sure that you
provide the GitHub personal access token!

At Timber, we tag releases before pushing them to GitHub. We then use `grease
update-release` to flesh out the release. The resulting script looks something
like this:

```shell
grease update-release \
  --name $name \
  --notes $notes \
  --assets "dist/*" \
  timberio/grease \
  $tag
```

You can check out the Makefile's "release" goal for the real deal.

### Uploading Assets to a Release

If you have an existing release and only want to add assets to it, you use
the `upload-assets` sub-command which takes three positional arguments:
the repository name, the release tag, and a glob pattern used to find files.

```
grease upload-assets timberio/grease v1.0.0 "dist/*"
```

The only additional flag this sub-commmand accepts is `--github-token` which you
can also pass in via the `GIHUB_TOKEN` environment variable. The GitHub personal
access token is required in order to upload the assets.

### Listing Files Matching Glob Pattern

To check which files will match a glob pattern, you can use the `list-files`
sub-command which takes one positional argument: the glob pattern.

```shell
grease list-files "dist/*"
```

This sub-command does _not_ require a GitHub personal access token to be provided
(since it doesn't do anything with GitHub).

It also doesn't accept any additional flags.

### Additional Help

You can use the `help` sub-command to get built-in help from Grease. Just follow
the sub-command with the name of another sub-command:

```shell
grease help update-release
```

## License & Copyright

Grease is copyright 2017 by Timber Technologies, Inc. and made available under
the ISC License. See the LICENSE file in the code repository or included in
distribution archives for full details.

[![Built by Timber.io](https://res.cloudinary.com/timber/image/upload/v1503615886/built_by_timber_wide.png)](https://timber.io/?utm_source=github&utm_campaign=timberio%2Fgrease)
