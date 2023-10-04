# github-release-stats

This is a small tool that queries repositories in GitHub for statistics about x.y.0 releases:

- number of x.y.0 releases
- minimum number of days between x.y.0 releases
- average number of days between x.y.0 releases
- maximum number of days between x.y.0 releases
- standard deviation

## GitHub token configuration

You'll need to create a [GitHub token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) to use this tool.

Once you've created a token, set the `GITHUB_TOKEN` environment variable. Now you're ready to go!

## Usage

```shell
Usage of github-release-stats:
      --owner string    Default owner to use for repos not in the OWNER/NAME format.
      --repos strings   List of repos to query. May specify as NAME or OWNER/NAME. If OWNER is omitted, falls back to --owner.
```

List stats for multiple repositories in the same organization:
```shell
$ github-release-stats --owner actions --repos runner,labeler,add-to-project
2023/10/04 16:05:15 Getting stats for actions/runner
2023/10/04 16:05:17 Getting stats for actions/labeler
2023/10/04 16:05:17 Getting stats for actions/add-to-project
Repo                    x.y.0 releases  Min days between  Avg days between  Max days between  StdDev
actions/runner          56              4.79              28.18             96.00             17.20
actions/labeler         5               10.99             286.36            921.19            361.03
actions/add-to-project  5               1.18              73.99             132.91            52.83 
```

List stats for multiple repositories in different organizations:
```shell
$ github-release-stats  --repos kubernetes/kubernetes,actions/runner
2023/10/04 16:06:54 Getting stats for kubernetes/kubernetes
2023/10/04 16:06:58 Getting stats for actions/runner
Repo                   x.y.0 releases  Min days between  Avg days between  Max days between  StdDev
kubernetes/kubernetes  33              13.75             97.78             291.25            47.62
actions/runner         56              4.79              28.18             96.00             17.20
```

## License

This project is licensed under the [Apache 2.0 License](LICENSE).

## Contributing

Please see the [contributing guide](CONTRIBUTING.md) for details on contributing to this project.
