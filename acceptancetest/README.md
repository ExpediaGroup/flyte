# Acceptance tests for flyte-api

To run acceptance tests:
  

```
go test ./... -tags=acceptance
```

These will start a disposable docker mongo container and flyte-api on randomly available TCP ports. If mongo can't be started (e.g docker is not available in the path),
tests will be skipped and won't fail the build.

Acceptance tests will not run when building flyte-api in docker.

The tests can be run in an IDE by running the test suite in "acceptance_test.go".