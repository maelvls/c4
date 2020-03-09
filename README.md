# Nuke Cloud

We keep having many leftover VMs when running integration tests. `c4` aims
at removing anything that costs $$. Claudia wanted to call it
`small-spender` but I went to fast and called it `c4`.

- **Why not use [cloud-nuke](https://github.com/gruntwork-io/cloud-nuke)?**
  That's because they do much more than deleting VMs and SGs, and they are
  specific to AWS.
- **Why not duct-tape a bash script?** Because we want a `--older-than 1h`
  flag, and doing that through scripting is a lot of pain.
- Why the hell did you use environment variables? Env vars are bad and hard
  to discover! That's because our current local and CI environment have all
  these env variables through `source .passrc`. And to remediate the issue
  of discoverability, I did two things:
  - environment variables must be not empty, so it's easy not to miss
      any,
  - `--help` shows all the env vars with their description.

`c4` uses enviroment variables for reading AWS, GCP and OpenStack
credentials and then proceeds we cleaning up all VMs and SGs. Details about
these env vars is available with `--help`.

## Usage

```sh
% go install github.com/ori-edge/c4
% source .passrc
% c4 --aws-name-contains="-test-" --os-name-contains="-test-"

Removing anything older than 24h0m0s.
Note: running in dry-mode. To actually delete things, add --do-it.
found aws instance test-aws-machine-crwvq (i-0d2efdf6b74578413), removing since age is 171h17m33.371533s
found aws instance test-machine-225dj (i-0e23754b0de933a8e), removing since age is 216h0m13.371556s
found aws instance test-aws-machine-2x48p (i-0d85e47e2e029b479), removing since age is 192h39m34.371561s
found aws instance test-aws-machine-72nmq (i-08c534433d1557efd), removing since age is 215h48m52.371565s
found aws instance test-aws-machine-mw7fr (i-0aa7b5c6805d22fd1), removing since age is 188h10m19.371569s
found openstack instance server-test-1 (39aa7efe-ff18-4144-9d62-d4cba84dbd47), keeping it since age is 2m37.175715s
```

Check visually that everything looks good and re-run with `--do-it` to
actually delete them.

## Help

```sh
% c4 --help
Usage of c4:
  -aws-name-contains string
    	selects AWS instances where tag:Name contains this string
  -do-it
    	By default, nothing is deleted. This flag enable deletion.
  -older-than duration
    	Only delete resources older than this specified value. Can be any valid Go duration, such as 10m or 8h. (default 24h0m0s)
  -os-name-contains string
    	selects OpenStack instances where the instance name contains this string

Mandatory environment variables:
  AWS_ACCESS_KEY_ID
    	The AWS access key
  AWS_SECRET_ACCESS_KEY
    	The AWS secret key
  AWS_REGION
    	The AWS region
  OS_USERNAME

  OS_PASSWORD

  OS_AUTH_URL
    	looks like http://host/identity/v3
  OS_PROJECT_NAME
    	Also called 'tenant name'
  OS_REGION
    	e.g., UK1 (for OVH)
  OS_PROJECT_DOMAIN_NAME
    	that's "Default" for most OpenStack instances
```
