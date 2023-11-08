# pull-request-title-validation

The `pull-request-title-validation` GitHub Action helps ensuring that contributors provide informative and well-formatted titles - based on the [conventional-commits] specification. The titles of the pull request could then be used to create automated releases.

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
  validation:
    name: validate-pull-request-title
    runs-on: ubuntu-latest
    steps:
      - name: validate pull request title
        uses: kontrolplane/pull-request-title-validation@v1
```
