#!/bin/sh -eu
echo "Environment:"
env

echo "Contents of GITHUB_WORKSPACE:"
ls -l $GITHUB_WORKSPACE
