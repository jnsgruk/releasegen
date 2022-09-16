# releasegen

This is a tool used for generating JSON reports containing details about Github and Launchpad
version control repositories.

## Why?

I lead some teams at Canonical, and I wanted a way to track the Github Releases of various teams
over time, and provide a single place for different departments, partners and customers to see new
releases of my teams' work.

The result of this is: https://jnsgruk.github.io/releases

This tool is used to generate a static JSON file every few minutes on a timer, which is then used
to generate the static site.

## Usage

```
releasegen is a utility for enumerating Github and Launchpad releases/tags from
specified Github Organisations or Launchpad project groups.

This tool is configured using a single file in one of the three following locations:

        - ./releasegen.yaml
        - $HOME/.config/releasegen.yaml
        - /etc/releasegen/releasegen.yaml

For more details on the configuration format, see the homepage below.

Prior to launching, you must also set an environment variable named RELEASEGEN_TOKEN whose
contents is a Github Personal Access token with sufficient rights over any org you wish to
query.

For example:

        export RELEASEGEN_TOKEN=ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ

You can create a Personal Access Token at: https://github.com/settings/tokens

Homepage: https://github.com/jnsgruk/releasegen

Usage:
  releasegen [flags]

Flags:
  -h, --help      help for releasegen
  -v, --version   version for releasegen
```

## Configuration Format

The tool is configured with a simple YAML file named `releasegen.yaml`. This file can be in one of
the following places:

- `$(pwd)/releasegen.yaml`
- `$HOME/.config/releasegen.yaml`
- `/etc/releasegen/releasegen.yaml`

You can find an [example config file](./releasegen.yaml.example) in this repository.

The configuration file format is as follows:

```yaml
# (Required) A list of teams to gather information for
teams:
  # (Required) The name of a real-life team
  - name: <team name>

    # (Optional) A list of Github org configurations for the team
    github:
      # (Required): The name of a Github Organisation
      - org: <github organisation name>

        # (Required) A list of teams to query from the Github Org
        teams:
          # The slug name of the Github org
          - <team>
          - <team>
          - ...

        # (Optional) A list of repository names to ignore
        ignores:
          # List of repo names
          - <repo>
          - <repo>

    # (Optional) Launchpad configuration for the team
    launchpad:
      # (Required) A list of Launchpad Project Groups to query
      project-groups:
        - <project group>
        - <project group>
```

## Development

This project uses [goreleaser](https://goreleaser.com/) to build and release.

You can get started by just using Go, or with goreleaser:

```shell
# Clone the repository
git clone https://github.com/jnsgruk/releasegen
cd releasegen

# Build/run with Go
go run main.go

# Build a snapshot release with goreleaser (putput in ./dist)
goreleaser build --rm-dist --snapshot
```

## Contributing / TODO

Contributions are welcome through the means of issues, pull requests or whatever, really.

Some things I'd like to get around to:

- [ ] Add some unit tests (!!!)
- [ ] Reconsider output format
  - `body_html` -> `body`
  - `href` -> `release_url`
  - `compare_href` -> `compare_url`
