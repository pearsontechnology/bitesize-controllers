#!/bin/bash

set -x

DRY_RUN=false

VERSION_SOURCE_FILE="controller.go"
REPOSITORY="pearsontechnology/nginx-ingress"

BUILD=true

FORCE=false
PUBLISH=false

MAJOR=false
MINOR=false
PATCH=false
UPREV=false

VERSION_LINE=`grep -oEi "const[\t ]+version[\t ]*=[\t ]*\"[0-9]+\.[0-9]+\.[0-9]+\"" $VERSION_SOURCE_FILE`
VERSION=`echo $VERSION_LINE | grep -oEi "[0-9]+\.[0-9]+\.[0-9]+"`
CODE_VERSION=$VERSION
PUBLISHED_TAGS=`wget -q https://registry.hub.docker.com/v1/repositories/$REPOSITORY/tags -O - | jq ".[].name" -r`
PUBLISHED_VERSIONS=`echo $PUBLISHED_TAGS | grep -oEi "[0-9]+\.[0-9]+\.[0-9]+"`
TAG=$VERSION

Showhelp () {
  echo "build.sh <options>"
  exit 0
}

while [[ $# > 0 ]]
do
  key="$1"
  case $key in
    --publish|-P)
      PUBLISH=true
      ;;
    --force|-f)
      FORCE=true
      ;;
    --tag|-t)
      TAG="$2"
      shift
      ;;
    --version|-v)
      if [[ "$TAG" == "$VERSION" ]]; then
        TAG="$2"
      fi
      VERSION="$2"
      shift
      ;;
    -M|--major)
      MAJOR=true
      UPREV=true
      ;;
    -m|--minor)
      MINOR=true
      UPREV=true
      ;;
    -p|--patch)
      PATCH=true
      UPREV=true
      ;;
    -s|--source-file)
      VERSION_SOURCE_FILE="$2"
      shift
      ;;
    --no-build)
      BUILD=false
      ;;
    --dry-run)
      DRY_RUN=true
      ;;
    -\?|-h|--help)
      Showhelp
      ;;
    *)
    # unknown option
    ;;
  esac
  shift # past argument or value
done

if [[ $UPREV == true ]]; then
  a=( ${VERSION//./ } )
  if [[ $MAJOR == true ]]; then
    ((a[0]++))
    a[1]=0
    a[2]=0
  fi

  if [[ $MINOR == true ]]; then
    ((a[1]++))
    a[2]=0
  fi

  if [[ $PATCH == true ]]; then
    ((a[2]++))
  fi
  NEW_VERSION="${a[0]}.${a[1]}.${a[2]}"
  if [[ "$TAG" == "$VERSION" ]]; then
    TAG=$NEW_VERSION
  fi
  VERSION=$NEW_VERSION
fi

if [[ $PUBLISH ]]; then
  if [[ $FORCE != true ]]; then
    if [[ $VERSION == $TAG ]]; then
      CHECK=`echo $PUBLISHED_VERSIONS | grep -o "$VERSION"`
      if [[ "$CHECK" != "" ]]; then
        echo "Version $VERSION already exists in the registry!"
        exit 1
      fi
    fi
    CHECK=`echo $PUBLISHED_TAGS | grep -o "$TAG"`
    if [[ "$CHECK" != "" ]]; then
      echo "Tag $TAG already exists in the registry!"
      exit 1
    fi
  fi
fi

if [[ "$CODE_VERSION" != "$VERSION" ]]; then
  NEW_VERSION_OK=`echo $VERSION | grep  -oEi "[0-9]+\.[0-9]+\.[0-9]+"`
  if [[ "$NEW_VERSION_OK" == "" ]]; then
    echo "Invalid new version specified: $VERSION"
    exit 1
  fi
  NEW_VERSION_LINE="const version = \"$VERSION\""
  if [[ $DRY_RUN == true ]]; then
    echo sed -i "s/$VERSION_LINE/$NEW_VERSION_LINE/" $VERSION_SOURCE_FILE
  else
    sed -i "s/$VERSION_LINE/$NEW_VERSION_LINE/" $VERSION_SOURCE_FILE
  fi
fi

if [[ $BUILD == true ]]; then
  echo "Building version: "$VERSION

  if [[ $DRY_RUN == true ]]; then
    echo go build "$VERSION_SOURCE_FILE"
  else
    go build "$VERSION_SOURCE_FILE"
  fi
fi

if [[ $PUBLISH == true ]]; then
  echo "Publishing $VERSION"
  if [[ $DRY_RUN == true ]]; then
    echo docker build -t $REPOSITORY:$TAG .
    echo docker push $REPOSITORY:$TAG
  else
    docker build -t $REPOSITORY:$TAG .
    docker push $REPOSITORY:$TAG
  fi
fi
