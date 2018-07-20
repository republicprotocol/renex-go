#!/bin/bash

MODULE_FOLDER="modules/renex-js"
UI_FOLDER="ui"

BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Add modules
if [ -d $MODULE_FOLDER ]; then
    cd $MODULE_FOLDER
    git checkout $BRANCH
    git pull
    cd ../..
else
    git clone -b $BRANCH https://github.com/republicprotocol/renex-js.git modules/renex-js
fi

# Remove the old build folder
rm -rf $UI_FOLDER

# Build UI
cd $MODULE_FOLDER
npm install
npm run build
mv build ../../$UI_FOLDER
cd ../..
