# C4, the VM annihilator

[![CI](https://github.com/maelvls/c4/workflows/CI/badge.svg)](https://github.com/maelvls/c4/actions)

## Download and install

Get the [binary](https://github.com/maelvls/c4/releases) or go get it
with:

```sh
go install github.com/maelvls/c4@latest
```

You can use `c4` to monitor all the VMs that are created for testing
purposes using `watch --color c4`:

![Demo of c4. This gif is hosted in the description of PR #1](https://user-images.githubusercontent.com/2195781/77187108-3ef3d680-6ad4-11ea-986a-eff98cda5b20.gif)

## Why c4

We keep having many leftover VMs when running integration tests. `c4` aims
at removing anything that costs $$. Claudia wanted to call it
`small-spender` but I went to fast and called it `c4`.

- **Why not use [cloud-nuke](https://github.com/gruntwork-io/cloud-nuke)?**
  That's because they do much more than deleting VMs and SGs, and they are
  specific to AWS.
- **Why not duct-tape a bash script?** Because we want a `--older-than 1h`
  flag, and doing that through scripting is a lot of pain.
- **Why the hell did you use environment variables?** Env vars are bad and
  hard to discover! That's because it allows me to use `.envrc` (direnv)
  locally. To remediate the issue of discoverability, I did two things:
  - environment variables must be not empty, so it's easy not to miss any,
  - `--help` shows all the env vars with their description.

Important: for AWS and OpenStack, it will only focus on one specific
region. For GCP, it will work on all regions simultanously.

## How is it automated?

The `c4` binary is built by Github Actions and uploaded as a Github
[release](https://github.com/maelvls/c4/releases) using
[goreleaser](https://github.com/goreleaser/goreleaser).

You can tell `c4` to look for VMs names with a regex `(test|example)` and
with `--older-than 1h` so that only "old" VMs get deleted. Whenever a VM
gets deleted, `c4` is able to tell you through Slack. For example:

## Usage

Install `v4`:

```sh
go install github.com/maelvls/c4@latest
```

First, set a couple of variables (`direnv` is recommended):

```sh
export GCP_JSON_KEY=...
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export OS_USERNAME=...
export OS_PASSWORD=...
export SLACK_TOKEN=...
```

Run it to see the VMs that were found:

```sh
c4 --aws-regex="(test|example)" --os-regex="(test|example)" --gcp-regex="(test|example)" --older-than=1h --slack-channel=test-channel
Removing anything older than 24h0m0s.
Note: running in dry-mode. To actually delete things, add --do-it.
found aws instance test-aws-machine-crwvq (i-0d2efdf6b74578413), removing since age is 171h17m33.371533s
found aws instance test-machine-225dj (i-0e23754b0de933a8e), removing since age is 216h0m13.371556s
found aws instance test-aws-machine-2x48p (i-0d85e47e2e029b479), removing since age is 192h39m34.371561s
found aws instance test-aws-machine-72nmq (i-08c534433d1557efd), removing since age is 215h48m52.371565s
found aws instance test-aws-machine-mw7fr (i-0aa7b5c6805d22fd1), removing since age is 188h10m19.371569s
found openstack instance server-test-1 (39aa7efe-ff18-4144-9d62-d4cba84dbd47), keeping it since age is 2m37.175715s
```

Then, if you are OK with this list, you can add `--do-it` to the command to
actually delete the VMs:

```sh
c4 --aws-regex="(test|example)" --os-regex="(test|example)" --gcp-regex="(test|example)" --do-it --older-than=1h --slack-channel=test-channel
Removing anything older than 24h0m0s.
found aws instance test-aws-machine-crwvq (i-0d2efdf6b74578413), removing since age is 171h17m33.371533s
found aws instance test-machine-225dj (i-0e23754b0de933a8e), removing since age is 216h0m13.371556s
found aws instance test-aws-machine-2x48p (i-0d85e47e2e029b479), removing since age is 192h39m34.371561s
found aws instance test-aws-machine-72nmq (i-08c534433d1557efd), removing since age is 215h48m52.371565s
found aws instance test-aws-machine-mw7fr (i-0aa7b5c6805d22fd1), removing since age is 188h10m19.371569s
found openstack instance server-test-1 (39aa7efe-ff18-4144-9d62-d4cba84dbd47), keeping it since age is 2m37.175715s
```

## Help

```sh
% c4 --help
Usage of c4:
  -aws-regex string
    	Selects AWS instances where tag:Name contains this string. Example: (test|example) (default ".*")
  -do-it
    	By default, nothing is deleted. This flag enable deletion.
  -gcp-regex string
    	Selects OpenStack instances where the instance name contains this string. Example: (test|example) (default ".*")
  -older-than duration
    	Only delete resources older than this specified value. Can be any valid Go duration, such as 10m or 8h. (default 24h0m0s)
  -os-regex string
    	Selects OpenStack instances where the instance name contains this string. Example: (test|example) (default ".*")
  -slack-channel string
    	With this argument, c4 sends a message to this channel whenever VMs are deleted (doesn't send anything when this flag isn't passed). Requires SLACK_TOKEN to be set.
  -version
    	Watch out, returns 'n/a (commit none, built on unknown)' when built with 'go get'.

Environment variables:
  AWS_ACCESS_KEY_ID
    	The AWS access key. (mandatory)
  AWS_SECRET_ACCESS_KEY
    	The AWS secret key. (mandatory)
  AWS_REGION
    	The AWS region. (mandatory)
  OS_USERNAME
    	 (mandatory)
  OS_PASSWORD
    	 (mandatory)
  OS_AUTH_URL
    	Often looks like http://host/identity/v3. (mandatory)
  OS_PROJECT_NAME
    	Also called 'tenant name'. (mandatory)
  OS_REGION
    	E.g., UK1 (for OVH). (mandatory)
  OS_PROJECT_DOMAIN_NAME
    	That's "Default" for most OpenStack instances. (mandatory)
  GCP_JSON_KEY
    	The content of the json key in plain text, not base-64 encoded. (mandatory)
  SLACK_TOKEN
    	Slack OAuth token, create one at https://api.slack.com/apps.
```

## Slack

Optionally, you can set `SLACK_TOKEN` and `--slack-channel` to send a Slack
message that sums up all the VMs that were deleted:

- create a [Slack App](https://api.slack.com/apps/) with the name `c4` (for
  example),
- add the Bot Token Scope `chat:write`,
- Copy the OAuth Access Token and use it as `SLACK_TOKEN`,
- Add `c4` to the channel you want the bot to be sending messages to.

## Changelog

## v1.1.0 - 04 July 2021

You can now use flags to set most values. For example, `GCP_JSON_KEY` can now be
set using the flag `--gcp-json-key`.

Unrelated but here are the commands you can use to create your GCP JSON key and
store it in [`lpass`](https://github.com/lastpass/lastpass-cli) (LastPass's
CLI):

```sh
gcloud iam service-accounts create c4-cli --project $(gcloud config get-value project | tr ':' '/')
gcloud projects add-iam-policy-binding jetstack-mael-valais --member=serviceAccount:c4-cli@$(gcloud config get-value project | tr ':' '/').iam.gserviceaccount.com --role=roles/compute.admin
lpass edit c4-cli --non-interactive <<EOF
Password: $(gcloud iam service-accounts keys create /dev/stdout --iam-account c4-cli@$(gcloud config get-value project | tr ':' '/').iam.gserviceaccount.com | jq -c)
EOF
```

Then you can run:

```sh
c4 --gcp-json-key="$(lpass show c4-cli -p)"
```

## v1.0.2 - 11 March 2019

- Bug: the Slack message was mixing up AWS and OpenStack, it now properly
  shows what is what.

## v1.0.1 - 9 March 2019

- Bug: fix a bug where a Slack message was sent even when no VM had been
  deleted.

## v1.0.0 - 9 March 2019

- Feature: use `--gcp-regex`, `--aws-regex` and `--os-regex` to filter
  which VMs should be removed. You can test the regex using
  <https://regex101.com> (flavor: Golang).
- Feature: to actually delete VMs, use `--do-it`. By default, it will run
  in dry-run mode.
- Feature: use `SLACK_TOKEN` and `--slack-channel` to let your team know
  which VMs have been wiped.
- Feature: credentials are passed through env variables. To list them, use
  `--help`.
