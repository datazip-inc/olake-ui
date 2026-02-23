#!/bin/sh
set +e  # Continue on errors

COLOR_BLUE="\033[0;94m"
COLOR_GREEN="\033[0;92m"
COLOR_RESET="\033[0m"

echo -e "${COLOR_BLUE}Setting up OLake UI Dev Container...${COLOR_RESET}"

###############################################################################
# Fix Go Version Mismatch (Force Install Go 1.24.13)
###############################################################################
REQUIRED_GO_VERSION="1.24.13"

# Enable Go auto toolchain download
export GOTOOLCHAIN=auto

CURRENT_GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')

if [ "$CURRENT_GO_VERSION" != "$REQUIRED_GO_VERSION" ]; then
  echo -e "${COLOR_BLUE}Installing Go ${REQUIRED_GO_VERSION} (current: ${CURRENT_GO_VERSION})...${COLOR_RESET}"
  apk add --no-cache wget tar > /dev/null 2>&1
  wget -q https://go.dev/dl/go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
  rm go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
else
  echo -e "${COLOR_GREEN}Go ${REQUIRED_GO_VERSION} already installed.${COLOR_RESET}"
fi

export PATH=/usr/local/go/bin:$PATH

echo -e "${COLOR_GREEN}Using Go version:${COLOR_RESET}"
go version

###############################################################################
# Install required system dependencies
###############################################################################
echo -e "${COLOR_BLUE}Installing system dependencies (git, node, npm, docker-cli)...${COLOR_RESET}"
apk add --no-cache git nodejs npm docker-cli > /dev/null 2>&1

echo -e "${COLOR_BLUE}Installing pnpm...${COLOR_RESET}"
npm install -g pnpm > /dev/null 2>&1

###############################################################################
# Backend Setup (Like Dockerfile Go Builder Stage)
###############################################################################
echo -e "${COLOR_BLUE}Downloading Go modules...${COLOR_RESET}"
cd /app/server || exit 1
go mod download

echo -e "${COLOR_BLUE}Building Go backend (olake-server)...${COLOR_RESET}"
go build -ldflags="-w -s" -o /app/olake-server .

###############################################################################
# Frontend Setup (Like Dockerfile Node Builder Stage)
###############################################################################
echo -e "${COLOR_BLUE}Installing UI dependencies (pnpm install)...${COLOR_RESET}"
cd /app/ui || exit 1
pnpm install

echo -e "${COLOR_BLUE}Building UI (pnpm build)...${COLOR_RESET}"
pnpm build

###############################################################################
# Banner
###############################################################################
echo -e "${COLOR_BLUE}
     %########%      
     %###########%       ____                 _____                      
         %#########%    |  _ \   ___ __   __ / ___/  ____    ____   ____ ___ 
         %#########%    | | | | / _ \\\\ \\ / / \\___ \\ |  _ \\  / _  | / __// _ \\\\
     %#############%    | |_| |(  __/  \\ V /  ____) )| |_) )( (_| |( (__(  __/
     %#############%    |____/  \\___|   \\_/   \\____/ |  __/  \\__,_| \\___\\\\\\___|
 %###############%                                  |_|
 %###########%${COLOR_RESET}

Welcome to your ${COLOR_GREEN}OLake UI${COLOR_RESET} development container!

Build completed:
- Backend binary: ${COLOR_GREEN}/app/olake-server${COLOR_RESET}
- Frontend build:  ${COLOR_GREEN}/app/ui/dist${COLOR_RESET}

Run manually:
- ${COLOR_GREEN}cd /app/server && go run .${COLOR_RESET}
- ${COLOR_GREEN}cd /app/ui && pnpm dev${COLOR_RESET}
- ${COLOR_GREEN}/app/olake-server${COLOR_RESET}

Backend serves on port ${COLOR_GREEN}8000${COLOR_RESET}

To enter this UI pod from local machine:
${COLOR_GREEN}devspace enter --label-selector app.kubernetes.io/name=olake-ui -n olake -c olake-ui${COLOR_RESET}
"

###############################################################################
# Shell Setup
###############################################################################
export PS1="\[${COLOR_BLUE}\]devspace-ui\[${COLOR_RESET}\] ./\W \[${COLOR_BLUE}\]\\$\[${COLOR_RESET}\] "
if [ -z "$BASH" ]; then export PS1="$ "; fi

export PATH="./bin:$PATH"

sh
