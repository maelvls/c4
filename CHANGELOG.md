# Changelog

See "[keep a Changelog](https://keepachangelog.com/en/1.0.0/)". This
project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

## v1.0.0 (to be released)

- Use `--gcp-name-contains`, `--aws-name-contains` and `--os-name-contains`
  to filter which VMs should be removed.
- To actually delete VMs, use `--do-it`. By default, it will run in dry-run
  mode.
- Credentials are passed through env variables. To list them, use `--help`.
