#!/bin/bash

MODULE_FOLDER="modules/renex-js"
UI_FOLDER="ui"

# Add modules
if [ -d "$MODULE_FOLDER" ]; then
    cd "$MODULE_FOLDER"
    git pull
    cd ../..
else
    # TODO: Use `master` branch once merged
    git clone -b template https://github.com/republicprotocol/renex-js.git modules/renex-js
fi

# Remove the old build folder
rm -rf $UI_FOLDER

# Build UI
cd "$MODULE_FOLDER"
npm install
npm run build
mv build "../../$UI_FOLDER"
cd ../..
