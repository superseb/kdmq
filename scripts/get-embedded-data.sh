#!/usr/bin/env bash

# List releases
for release in $(gh api graphql -F owner='rancher' -F name='rancher' -f query='query($name: String!, $owner: String!) {repository(owner: $owner, name: $name) {releases(last: 30) {nodes { tagName }}}}' | jq -r .data.repository.releases.nodes[].tagName); do
  # Broken releases without assets
  if [[ "${release}" == "v2.5.12-rc3" || "${release}" == "v2.6.5-alpha1" || "${release}" == "v2.6.4-alpha1" || "${release}" == "v2.6.4-alpha2" ]]; then
    continue
  fi
  # Only add if not already exists
  if [ ! -f "./embedded/data.${release}.json" ]; then
    docker pull -q rancher/rancher:$release &> /dev/null
    DID=$(docker create rancher/rancher:${release})
    docker cp $DID:/var/lib/rancher-data/driver-metadata/data.json ./embedded/data.$release.json
    echo "- Added ${release}"
  fi
done
