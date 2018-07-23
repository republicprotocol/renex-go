#!/bin/bash

MODULE_FOLDER="modules/renex-js"
UI_FOLDER="ui"

if [ "$1" != "--branch" ] || [ "$2" = "" ]; then
    echo "Please specify a branch to build using the --branch flag"
    exit 1
fi

BRANCH=$2

# Add modules
if [ -d $MODULE_FOLDER ]; then
    cd $MODULE_FOLDER
    git checkout $BRANCH
    git pull origin $BRANCH
    cd ../..
else
    git clone -b $BRANCH git@github.com:republicprotocol/renex-js.git "$MODULE_FOLDER"
fi

cd "$MODULE_FOLDER"
LATEST_COMMIT="`git rev-parse HEAD`"
cd ../..

echo -n "$LATEST_COMMIT" > env/latest_commit.txt


# Remove the old build folder
rm -rf $UI_FOLDER

# Build UI
cd $MODULE_FOLDER
npm install
npm run build
mv build ../../$UI_FOLDER
cd ../..
