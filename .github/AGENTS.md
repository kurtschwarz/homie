# GitHub Actions

- Pin every action in `.github/workflows/` to a full-length commit SHA, not a tag or branch. The repository rejects `uses: actions/checkout@v4` and fails the run.
- Put the version in a trailing comment, e.g. `uses: actions/checkout@<sha> # v4.4.0`.
- Find the SHA with `git ls-remote https://github.com/<owner>/<repo> 'refs/tags/<tag>*'`. If the tag is annotated, use the `^{}` line, which is the commit; the plain line is the tag object.
