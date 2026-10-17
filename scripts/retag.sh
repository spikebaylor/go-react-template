#!/usr/bin/env bash
# retag.sh FROM TAGS: give each of this repo's images tagged FROM the
# comma-separated TAGS too, in the registry, without pulling or rebuilding.
# The new tags point at the same multi-platform manifest, digest for digest.
#
# IMAGE (as in docker-bake.hcl) names the images; -api and -web are derived.
set -euo pipefail

from=$1
IFS=, read -ra tags <<<"$2"
: "${IMAGE:?set IMAGE to the image name, e.g. ghcr.io/owner/repo}"

for image in "$IMAGE" "$IMAGE-api" "$IMAGE-web"; do
  args=()
  for tag in "${tags[@]}"; do args+=(--tag "$image:$tag"); done
  docker buildx imagetools create "${args[@]}" "$image:$from"
done
