#!/bin/bash

# Print commands as they are executed
set -x

# Exit if any command fails
set -e

BASE_FOLDER=$(pwd)

RENEX_MODULE_FOLDER="$BASE_FOLDER/modules/renex-js"
SDK_MODULE_FOLDER="$BASE_FOLDER/modules/renex-sdk-ts"
UI_FOLDER="$BASE_FOLDER/ui"

if [ "$1" != "--branch" ] || [ "$2" = "" ]; then
    echo "Please specify a branch to build using the --branch flag"
    exit 1
fi

BRANCH=$2

# Add modules
if [ -d $RENEX_MODULE_FOLDER ]; then
    cd $RENEX_MODULE_FOLDER
    git checkout $BRANCH
    git pull origin $BRANCH
    cd $BASE_FOLDER
else
    git clone -b $BRANCH git@github.com:republicprotocol/renex-js.git "$RENEX_MODULE_FOLDER"
fi
if [ -d $SDK_MODULE_FOLDER ]; then
    cd $SDK_MODULE_FOLDER
    git checkout $BRANCH
    git pull origin $BRANCH
    cd $BASE_FOLDER
else
    git clone -b $BRANCH git@github.com:republicprotocol/renex-sdk-ts.git "$SDK_MODULE_FOLDER"
fi

# Get latest commit hash
cd "$RENEX_MODULE_FOLDER"
LATEST_COMMIT="`git rev-parse HEAD`"
cd $BASE_FOLDER
echo -n "$LATEST_COMMIT" > env/latest_commit.txt

# Remove the old build folder
rm -rf $UI_FOLDER

# Build SDK
cd $SDK_MODULE_FOLDER
npm install
npm run build:dev
cd $BASE_FOLDER

# Link UI and SDK and build UI
cd $RENEX_MODULE_FOLDER
npm install
cp -r $SDK_MODULE_FOLDER/lib ./node_modules/renex-sdk-ts
npm run build
cd $BASE_FOLDER
mv $RENEX_MODULE_FOLDER/build $UI_FOLDER
