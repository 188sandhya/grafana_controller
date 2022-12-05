#!/usr/bin/env bash
set -eo pipefail

PROJECT=eb-grafana-controller

BUILD_MODE="BUILD_MODE="
if [[ $* == *--blitz ]]; then
  BUILD_MODE="BUILD_MODE=blitz"
fi

if [[ $* == *--debug ]]; then
  BUILD_MODE="BUILD_MODE=debug"
fi

function main {
    echo "$@"
  case "${1:-}" in

  setup) setup;;
  minikube) minikube "${@:2}";;
  helm) helm "${@:2}";;
  build) build;;
  deploy-local) deployLocal;;
  cdc) cdc;;
  cdc-preserve) cdc-preserve;;
  *)
    help
    exit 1
    ;;

  esac
}

function help {
  echo "Usage:"
  echo " build                                   build docker image"
  echo " deploy-local                            build all containers and deploy them on minikibe"
  echo "                                              add --blitz to skip linter and unit tests"
  echo "                                              add --debug to build special for debugging"
  echo
  echo " cdc                                        runs the cdc test once"
  echo " cdc-preserve                               runs the cdc test once and leaves containers to dev"
  echo
}

function initDockerEnv {
  if [ -n "${DOCKER_HOST}" ]
  then
    echo "DOCKER_HOST already set"
  else
    eval $(minikube docker-env --shell bash)
    if [ -n "${DOCKER_CERT_PATH}" ] && [ "${DOCKER_CERT_PATH:0:1}" != '/' ]
    then
      DOCKER_CERT_PATH=$(wslpath -u "${DOCKER_CERT_PATH}")
    fi
  fi
}

function build {
  initDockerEnv
  case $BUILD_MODE in

    BUILD_MODE=blitz)
      skaffold run -f waas-config/environments/local/skaffold.yaml -p minikube-no-tests
      ;;

    BUILD_MODE=debug)
      skaffold run -f waas-config/environments/local/skaffold.yaml -p minikube-debug
      ;;

    *)
      skaffold run -f waas-config/environments/local/skaffold.yaml -p minikube --no-prune=false --cache-artifacts=false
      ;;
  esac
}

function deployLocal {
  case $BUILD_MODE in

    BUILD_MODE=blitz)
      skaffold run -f waas-config/environments/local/skaffold.yaml -p minikube-no-tests
      ;;

    BUILD_MODE=debug)
      skaffold run -f waas-config/environments/local/skaffold.yaml -p minikube-debug
      ;;

    *)
      skaffold run -f waas-config/environments/local/skaffold.yaml -p minikube --no-prune=false --cache-artifacts=false
      ;;
  esac
}

function cdc {
  docker-compose -f docker-compose-cdc.yml down
  buildCdcMock
  echoGreenText 'Running cdc tests...'
  docker-compose -f docker-compose-cdc.yml run executor
  docker-compose -f docker-compose-cdc.yml down 
  echoBlueText '... finished.'
}

function cdc-preserve {
  docker-compose -f docker-compose-cdc.yml down
  buildCdcMock
  echoGreenText 'Running cdc tests...'
  docker-compose -f docker-compose-cdc.yml run executor
  echoBlueText '... finished.'
}

function buildCdcMock {
  echoGreenText 'Building cdc mock containers...'
  docker build -t eb2/$PROJECT --build-arg "$BUILD_MODE" .
  docker build -t eb2/eb-initcruiser init-cruiser/.
  docker-compose -f docker-compose-cdc.yml build
}

function setup {
  echoGreenText 'Setup...'
  cd "$SOURCE_DIR"/grafana-controller 
  GO111MODULE=on go get github.com/swaggo/swag/cmd/swag@v1.6.3
  GO111MODULE=on go get github.com/onsi/ginkgo/ginkgo@v1.8.0
  GO111MODULE=on go get github.com/google/wire/cmd/wire@v0.4.0
  GO111MODULE=off go get -u golang.org/x/tools/cmd/stringer
  GO111MODULE=off go install github.com/golang/mock/mockgen
  go mod tidy
}

function deployLocalSkipTests {
  deployLocal "build-skip-tests"
}

function echoGreenText {
  if [[ "${TERM:-dumb}" == "dumb" ]]; then
    echo "${@}"
  else
    RESET=$(tput sgr0)
    GREEN=$(tput setaf 2)

    echo "${GREEN}${*}${RESET}"
  fi
}

function echoBlueText {
  if [[ "${TERM:-dumb}" == "dumb" ]]; then
    echo "${@}"
  else
    RESET=$(tput sgr0)
    BLUE=$(tput setaf 4)

    echo "${BLUE}${*}${RESET}"
  fi
}
function echoRedText {
  if [[ "${TERM:-dumb}" == "dumb" ]]; then
    echo "${@}"
  else
    RESET=$(tput sgr0)
    RED=$(tput setaf 1)

    echo "${RED}${*}${RESET}"
  fi
}

function echoWhiteText {
  if [[ "${TERM:-dumb}" == "dumb" ]]; then
     echo "${@}"
  else
    RESET=$(tput sgr0)
    WHITE=$(tput setaf 7)

    echo "${WHITE}${*}${RESET}"
  fi
}

main "$@"
