#!/bin/bash
# The script does automatic checking on a Go package and its sub-packages, including:
# 1. gofmt (http://golang.org/cmd/gofmt/)
# 2. go vet (http://golang.org/cmd/vet)
# 3. race detector (http://blog.golang.org/race-detector)
# 4. test coverage (http://blog.golang.org/cover)

set -e

# Automatic checks
test -z "$(gofmt -l -w .     | tee /dev/stderr)"
go vet ./...
go test -race ./...

# Run test coverage on each subdirectories and merge the coverage profile. 
echo "mode: count" > profile.cov
 
# Standard go tooling behavior is to ignore dirs with leading underscors
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d);
do
if ls $dir/*.go &> /dev/null; then
go test -covermode=count -coverprofile=$dir/profile.tmp $dir
if [ -f $dir/profile.tmp ]
then

cat $dir/profile.tmp | tail -n +2 >> profile.cov
rm $dir/profile.tmp
fi
fi
done
 
go tool cover -func profile.cov
