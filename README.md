# Nuke Cloud

We keep having many leftover VMs when running integration tests.
`nuke-clouds` aims at removing anything that costs $$.

- **Why not use gruntworks'
  [cloud-nuke](https://github.com/gruntwork-io/cloud-nuke)?** That's because
  they do much more than deleting VMs and SGs, and they are specific to
  AWS.
- **Why not duct-tape a bash script?** Because we want a `--older-than 1h`
  flag, and doing that through scripting is a lot of pain.

`nuke-clouds` uses enviroment variables for reading AWS, GCP and OpenStack
credentials and then proceeds we cleaning up all VMs and SGs. Details about
these env vars is available with `--help`.
