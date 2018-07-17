#!/bin/bash

# Add modules
rm -rf "./modules"
git clone -b template git@github.com:republicprotocol/renex-js.git modules/renex-js

# Remove the old build folder
rm -rf "./public"

# Build UI
cd ./modules/renex-js
npm install
npm run build
mv build "../../public"
cd ../..
