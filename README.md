# GitHub repositories secrets updater

Uses [GitHub Go SDK](https://github.com/octokit/go-sdk) to find repositories for an user authenticated with an access token, that are active and not forks, and update a secret if they already have a value for it.

## Inputs

Required options are taken from environment variables, as below. The environment variables are expected to be provided from a `.env` file.

* `AUTH_TOKEN`: the GitHub PAT with following permissions
  * "Metadata" repository permissions (read);
  * "Secrets" repository permissions (read and write).
* `SECRET_NAME`: the secret name;
* `SECRET_VALUE`: the secret value to set.
