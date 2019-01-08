#!/bin/bash

# Exit if any command fails
set -e

BASE_FOLDER=$(pwd)

RENEX_MODULE_FOLDER="$BASE_FOLDER/modules/renex-js"
SDK_MODULE_FOLDER="$BASE_FOLDER/modules/renex-sdk-ts"
UI_FOLDER="$BASE_FOLDER/ui"

RESET='\033[0m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'

usage() {
    echo "Please specify a network to deploy using -n"
    echo ""
    echo -e "  Available networks: ${BLUE}mainnet${RESET}, ${PURPLE}testnet${RESET}, ${CYAN}nightly${RESET}"
    echo ""
    exit 1
}

# Handle arguments
NOVERIFY=false
while getopts 'n:b::v' flag; do
  case "${flag}" in
    n) NETWORK=${OPTARG} ;;
    b) BRANCH=${OPTARG} ;;
    v) NOVERIFY=true ;;
    *) usage ;;
  esac
done

HEROKU_APP="renex-ui-$NETWORK"

if [ "$NETWORK" == "mainnet" ] && [ "$BRANCH" == "" ]; then
    BRANCH="master"
    COLOR="${BLUE}"
elif [ "$NETWORK" == "testnet" ] && [ "$BRANCH" == "" ]; then
    BRANCH="develop"
    COLOR="${PURPLE}"
elif [ "$NETWORK" == "nightly" ] && [ "$BRANCH" == "" ]; then
    BRANCH="nightly"
    COLOR="${CYAN}"
elif [ "$BRANCH" != "" ]; then
    COLOR="${CYAN}"
    heroku apps:create $HEROKU_APP
else
    usage
    exit 1
fi

echo -e "\nDeploying ${GREEN}renex-js:${BRANCH}${RESET} with ${GREEN}renex-sdk-ts:legacy${RESET} to ${COLOR}${NETWORK}${RESET}...\n"

# Print commands as they are executed
set -x

# Add modules
if [ -d $RENEX_MODULE_FOLDER ]; then
    cd $RENEX_MODULE_FOLDER
    # `npm install` may changes these files
    git checkout package.json package-lock.json
    git checkout $BRANCH
    git pull origin $BRANCH
    cd $BASE_FOLDER
else
    git clone -b $BRANCH git@github.com:republicprotocol/renex-js.git "$RENEX_MODULE_FOLDER"
fi
if [ -d $SDK_MODULE_FOLDER ]; then
    cd $SDK_MODULE_FOLDER
    # `npm install` may changes these files
    git checkout package.json package-lock.json
    git fetch
    git checkout legacy # TODO: Remove references to 'legacy' branch once renex-js is compatible with the new SDK
    git pull origin legacy
    cd $BASE_FOLDER
else
    git clone -b legacy git@github.com:republicprotocol/renex-sdk-ts.git "$SDK_MODULE_FOLDER"
fi

# Get latest renex-js commit hash and author
cd "$RENEX_MODULE_FOLDER"
LATEST_RENEX_COMMIT="`git rev-parse --short HEAD`"
LATEST_RENEX_AUTHOR="`git --no-pager  log --format='%aN <%aE>' HEAD^!`"
cd "$BASE_FOLDER"

# Get latest renex-sdk-ts commit hash
cd "$SDK_MODULE_FOLDER"
LATEST_SDK_COMMIT="`git rev-parse --short HEAD`"
LATEST_SDK_AUTHOR="`git --no-pager  log --format='%aN <%aE>' HEAD^!`"
cd "$BASE_FOLDER"

COMBINED_HASH="$LATEST_RENEX_COMMIT//$LATEST_SDK_COMMIT"

echo -n "${COMBINED_HASH}" > env/latest_commit.txt


# Remove the old build folder
rm -r $UI_FOLDER

# Build SDK
cd $SDK_MODULE_FOLDER
npm install
npm run build:dev
cd $BASE_FOLDER

# Link UI and SDK and build UI
cd $RENEX_MODULE_FOLDER
npm install
rm -r ./node_modules/renex-sdk-ts || true
cp -r $SDK_MODULE_FOLDER ./node_modules/renex-sdk-ts
npm run build
cd $BASE_FOLDER
mv $RENEX_MODULE_FOLDER/build $UI_FOLDER

set +x

echo -en "\n\n\n"
printf "${YELLOW}%`tput cols`s"|tr ' ' '='
echo -e "\r==== Git log details ${RESET}"
git status
printf "${YELLOW}%`tput cols`s${RESET}\n\n"|tr ' ' '='

echo "Version built from the following modules:"
echo -e "${GREEN}renex-js:${BRANCH}${RESET} [${YELLOW}${LATEST_RENEX_COMMIT}${RESET}] commited by ${YELLOW}${LATEST_RENEX_AUTHOR}${RESET}"
echo -e "${GREEN}renex-sdk-ts:legacy${RESET} [${YELLOW}${LATEST_SDK_COMMIT}${RESET}] commited by ${YELLOW}${LATEST_SDK_AUTHOR}${RESET}"
echo ""

if [ "$NETWORK" == "mainnet" ] || [ "$NOVERIFY" == false ]; then
    echo -en "Re-enter network name (${COLOR}${NETWORK}${RESET}) to confirm deploy: ${COLOR}"
    read CONFIRM
    echo -en "${RESET}"
    if [ "${CONFIRM}" != "${NETWORK}" ]; then
        echo -e "${RED}Mismatched network.${RESET}"
        exit 1
    fi
fi

set -x

git add ui env/latest_commit.txt
if [ "$(git diff --cached)" ]; then
    git commit -m "ui: built ${COMBINED_HASH}" --no-verify
    git push
fi

git add env
if [ "$(git diff --cached)" ]; then
    git commit -m "env: added '${NETWORK}' config" --no-verify
    git push
fi

# Push to Heroku
if [ -z "$(git config remote.heroku-${NETWORK}.url)" ]; then
    heroku git:remote -a ${HEROKU_APP}
    git remote rename heroku heroku-${NETWORK}
fi

git push "heroku-${NETWORK}" master

set +x

echo -e "\nPushed ${COMBINED_HASH} to ${COLOR}${NETWORK}${RESET}\n"
