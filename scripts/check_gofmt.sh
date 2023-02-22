#!/usr/bin/env bash

# Check go fmt
echo "==> Checking that go fmt was run for code ..."
gofmt_files=$(go fmt ./...)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "Use \`go fmt\` to reformat code."
    exit 1
fi

exit 0