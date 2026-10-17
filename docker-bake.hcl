# Builds the images from the Dockerfile's targets. `just image` builds them for
# this machine; CI sets these variables to build and push them for release.

# Image name, without a tag. The api and web images add -api and -web.
variable "IMAGE" { default = "app" }
# The version baked into the binary (/api/health reports it).
variable "VERSION" { default = "dev" }
# Comma-separated tags to apply to each image.
variable "TAGS" { default = "dev" }
# Comma-separated platforms; empty builds for this machine only.
variable "PLATFORMS" { default = "" }
# Where the source lives, for the image's OCI source label.
variable "SOURCE" { default = "" }

function "tags" {
  params = [suffix]
  result = [for t in split(",", TAGS) : "${IMAGE}${suffix}:${t}"]
}

target "_common" {
  args      = { VERSION = VERSION }
  platforms = PLATFORMS == "" ? [] : split(",", PLATFORMS)
  labels = {
    "org.opencontainers.image.version" = VERSION
    "org.opencontainers.image.source"  = SOURCE
  }
}

target "app" {
  inherits = ["_common"]
  target   = "combined"
  tags     = tags("")
}

target "api" {
  inherits = ["_common"]
  target   = "api"
  tags     = tags("-api")
}

target "web" {
  inherits = ["_common"]
  target   = "web"
  tags     = tags("-web")
}

group "default" {
  targets = ["app", "api", "web"]
}
