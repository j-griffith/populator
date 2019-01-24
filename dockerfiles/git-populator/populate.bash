#!/usr/bin/bash

REPO=$1
BRANCH=$2
DEST=$3
git clone -b "$BRANCH"  "$REPO" "$DEST"
