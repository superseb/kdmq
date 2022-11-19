#!/usr/bin/env bash

# List releases
for release in $(gh api graphql -F owner='rancher' -F name='rancher' -f query='query($name: String!, $owner: String!) {repository(owner: $owner, name: $name) {releases(last: 30) {nodes { tagName }}}}' | jq -r .data.repository.releases.nodes[].tagName | egrep "^v2.5|^v2.6|^v2.7"); do
  # Only add if not already exists
  if [ ! -f "./embedded/data.${release}.json" ]; then
    # we can just download the data from v2.7.0 and up
    if [[ "$release" == v2.7* ]]; then
      wget -q https://github.com/rancher/rancher/releases/download/$release/rancher-data.json -O ./embedded/data.$release.json
      echo "- Added ${release}"
    else
      docker pull -q rancher/rancher:$release &> /dev/null
      if [ $? -eq 0 ]; then
        DID=$(docker create rancher/rancher:${release})
        docker cp $DID:/var/lib/rancher-data/driver-metadata/data.json ./embedded/data.$release.json
        echo "- Added ${release}"
      fi
    fi
  fi
done
