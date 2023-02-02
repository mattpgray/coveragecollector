# coveragecollector

A simple go binary for collecting coverage information from multiple packages.

Can output:
- The total coverage across the entire codebase
- The coverage of each individual package

Other uses cases can be more easily achieved using the standard tools at the moment

## Usage

```
go install github.com/mattpgray/coveragecollector@latest
$(go env GOPATH)/bin/coveragecollector -total <coverage file> # Output the total coverage
$(go env GOPATH)/bin/coveragecollector -packages <coverage file> # Output the coverage for each package
```

## TODO

- Add support for combining multiple coverage files
