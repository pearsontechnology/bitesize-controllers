#!/bin/bash

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
TAG_LATEST=false
TAG_EXISTS=false

VERSION_LINE=`grep -oEi "const[\t ]+version[\t ]*=[\t ]*\"[0-9]+\.[0-9]+\.[0-9]+\"" $VERSION_SOURCE_FILE`
VERSION=`echo $VERSION_LINE | grep -oEi "[0-9]+\.[0-9]+\.[0-9]+"`
CODE_VERSION=$VERSION
PUBLISHED_TAGS=`wget -q https://registry.hub.docker.com/v1/repositories/$REPOSITORY/tags -O - | jq ".[].name" -r`
PUBLISHED_VERSIONS=`echo $PUBLISHED_TAGS | grep -oEi "[0-9]+\.[0-9]+\.[0-9]+"`
TAG=$VERSION
STABLE_NGINX_VERSION=`curl -s https://index.docker.io/v1/repositories/nginx/tags | jq -r '( .[].name )' | grep '^[0-9]*\.[0-9]*\.[0-9]*$' | sort -V | tail -1`
CURRENT_NGINX_VERSION=`grep FROM Dockerfile | awk -F ':' '{ print $2 }'`

function version_gt() {
  test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1";
}

Showhelp () {
  echo "build.sh <options>"
  echo ""
  echo "  Options"
  echo "    -? or -h or --help - Show this screen"
  echo "    -P or --publish - Publish the build to the docker registry"
  echo "    -f or --force - Force push the build to the docker registry even if it already exists"
  echo "    -t <tag> or --tag <tag> - What tag to use when pushing"
  echo "    -l or --tag-latest - Publish to docker registry as latest"
  echo "    -v <version> or --version <version> - Composite version number to deploy"
  echo "    -M or --major - Increment Major version number, reset minor and patch"
  echo "    -m or --minor - Increment Minor version number, reset patch"
  echo "    -p or --patch - Increment Patch version number"
  echo "    -s or --source-file <fileName> - Set the source file to work against, defaults to \"$VERSION_SOURCE_FILE\""
  echo "    --no-build - Don't execute go build"
  echo "    --dry-run - Perform a dry run, nothing will actually change, but all commands will be output"
  exit 0
}

if version_gt $STABLE_NGINX_VERSION $CURRENT_NGINX_VERSION; then
     echo -e "\033[33mNginx base image version $STABLE_NGINX_VERSION is available on Dockerhub. Please consider upgrading the ingress Dockerfile.\033[0m"
fi

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
    -l|--tag-latest)
      TAG_LATEST=true
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
  else # We want to force a push even if the tag exists
    if [[ $VERSION == $TAG ]]; then
      CHECK=`echo $PUBLISHED_VERSIONS | grep -o "$VERSION"`
      if [[ "$CHECK" != "" ]]; then
        TAG_EXISTS=true
      fi
    else
      CHECK=`echo $PUBLISHED_TAGS | grep -o "$TAG"`
      if [[ "$CHECK" != "" ]]; then
        TAG_EXISTS=true
      fi
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
    echo "Updating $VERSION_SOURCE_FILE with version $VERSION"
    sed -i "s/$VERSION_LINE/$NEW_VERSION_LINE/" $VERSION_SOURCE_FILE
    git add $VERSION_SOURCE_FILE
    git commit -m "v$VERSION"
    echo "Pushing version tag to repo"
    git push
  fi
fi

if [[ $BUILD == true ]]; then
  echo "Building version: "$VERSION

  if [[ $DRY_RUN == true ]]; then
    echo GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build "$VERSION_SOURCE_FILE"
  else
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build "$VERSION_SOURCE_FILE"
  fi
fi

if [[ $PUBLISH == true ]]; then
  echo "Publishing $VERSION as tag $TAG"
  if [[ $DRY_RUN == true ]]; then
    echo docker build -t $REPOSITORY:$TAG .
    echo docker push $REPOSITORY:$TAG
  else
    if [[ $TAG_EXISTS == true ]]; then
      docker rmi $REPOSITORY:$TAG
    fi
    docker build -t $REPOSITORY:$TAG .
    docker push $REPOSITORY:$TAG
    if [[ $TAG_LATEST == true ]]; then
      echo "Publishing $VERSION as tag latest"
      docker rmi $REPOSITORY:latest
      docker push $REPOSITORY:latest
    fi
  fi
fi
