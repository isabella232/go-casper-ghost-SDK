#!/bin/bash

VERSION=$1
TEST_TYPE=$2
BASE_PATH=./src/state_transition/spec_tests/.temp
REPO_NAME=eth2.0-spec-tests

# Remove dir if it already exists
rm -rf $REPO_NAME
mkdir $BASE_PATH

function download {
    TAR_FILE=$TEST_TYPE.tar.gz
    OUTPUT=$BASE_PATH/$TAR_FILE
    DOWNLOAD_URL=https://github.com/ethereum/$REPO_NAME/releases/download/$VERSION/$TAR_FILE
    wget $DOWNLOAD_URL -O $OUTPUT
    tar -xzf $OUTPUT -C $BASE_PATH
    rm $OUTPUT
}

download
