#!/bin/bash

MODULES_FOLDER="modules"
UI_FOLDER="ui"

# Add modules
if [ -d "$MODULES_FOLDER" ]; then
    cd "$MODULES_FOLDER/renex-js"
    git pull
    cd ../..
else
    # TODO: Use `master` branch once merged
    git clone -b template git@github.com:republicprotocol/renex-js.git modules/renex-js
fi

# Remove the old build folder
rm -rf $UI_FOLDER

# Build UI
cd "$MODULES_FOLDER/renex-js"
npm install
npm run build
mv build "../../$UI_FOLDER"
cd ../..
