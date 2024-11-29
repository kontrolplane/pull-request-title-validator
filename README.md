# pull-request-title-validator

The `pull-request-title-validator` GitHub Action helps ensuring that contributors provide informative and well-formatted titles - based on the [conventional-commits] specification. The titles of the pull request could then be used to create automated releases.

[conventional-commits]: https://www.conventionalcommits.org/en/v1.0.0/

## Example title

```
feat(client): add component
│    │        └─────── message
│    └──────────────── scope
└───────────────────── type
```

## Example usage

The action can be used with both the `pull_request` and `pull_request_target` trigger.

### Default

```yaml
name: validate-pull-request-title

on:
  pull_request:
    types:
      - opened
      - edited
      - synchronize

permissions:
  pull-requests: read

jobs:
  validator:
    name: validate-pull-request-title
    runs-on: ubuntu-latest
    steps:
      - name: validate pull request title
        uses: kontrolplane/pull-request-title-validator@v1.3.2
```

### Custom types

```yaml
name: validate-pull-request-title

on:
  pull_request:
    types:
      - opened
      - edited
      - synchronize

permissions:
  pull-requests: read

jobs:
  validator:
    name: validate-pull-request-title
    runs-on: ubuntu-latest
    steps:
      - name: validate pull request title
        uses: kontrolplane/pull-request-title-validator@v1.3.2
        with:
          types: "fix,feat,chore"
```

### Custom scopes

Scopes support regular expression patterns, allowing you to define specific patterns to match the scopes you want to allow. You can also separate multiple scopes using commas.

```yaml
name: validate-pull-request-title

on:
  pull_request:
    types:
      - opened
      - edited
      - synchronize

permissions:
  pull-requests: read

jobs:
  validator:
    name: validate-pull-request-title
    runs-on: ubuntu-latest
    steps:
      - name: validate pull request title
        uses: kontrolplane/pull-request-title-validator@v1.3.2
        with:
          scopes: "api,lang,parser,package/.+"
```

## contributors

[//]: kontrolplane/generate-contributors-list

<a href="https://github.com/levivannoort"><img src="https://avatars.githubusercontent.com/u/73097785?v=4" title="levivannoort" width="50" height="50"></a>
<a href="https://github.com/paopa"><img src="https://avatars.githubusercontent.com/u/52045032?v=4" title="paopa" width="50" height="50"></a>

[//]: kontrolplane/generate-contributors-list
