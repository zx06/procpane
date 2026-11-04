# SSH config test fixtures (upstream)

Vendored verbatim for procpane `internal/source/ssh` tests. Do not edit except to
refresh from upstream.

| Directory | Source | License / note |
|-----------|--------|----------------|
| `kevinburke-ssh_config/` | [kevinburke/ssh_config v1.4.0 `testdata/`](https://github.com/kevinburke/ssh_config/tree/v1.4.0/testdata) | MIT |
| `cocotyty-sshconf/` | [cocotyty/sshconf `test_cases/`](https://github.com/cocotyty/sshconf/tree/master/test_cases) | MIT |
| `openssh-regress/` | Excerpted `Host`/`ProxyJump`/`IdentitiesOnly`/`ConnectTimeout` stanzas from [OpenSSH `regress/`](https://github.com/openssh/openssh-portable/tree/master/regress) scripts (Public Domain) | Public Domain |

Glue in tests only: copy fixture into `$HOME/.ssh/config`, write Include targets
under `.ssh/` (same pattern as upstream `ssh_config` `TestInclude`), or point
`UserSettings.ConfigFinder` at a fixture path.
