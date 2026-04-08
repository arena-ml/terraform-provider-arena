# Copyright (c) ArenaML Labs Pvt Ltd.

terraform {
  required_providers {
    arenaml = {
      source = "arena-ml/arenaml"
    }
  }
}

provider "arenaml" {
  server_url = "http://localhost:18080/api/v1"
}


resource "arenaml_cluster_manager" "def" {
  name = "tf_test_engine"
  kind = "nomad"
  spec = file("${path.module}/default-engine-spec.json")
}

resource "arenaml_org" "nav_dev" {
  name        = "nav-dev"
  description = "developer group for navigation systems"
}

resource "arenaml_team" "auto_flight" {
  name        = "path-planning"
  role        = "devs"
  org_id      = arenaml_org.nav_dev.id
  description = "dev team working on autonomous path planning"
  config = {
    allow_cross_orgs = true
  }
}

resource "arenaml_user" "super_dev" {
  name  = "super-dev"
  email = "dev@arenaml.dev"
}


resource "arenaml_store" "ais_staging" {
  name     = "ais-staging"
  kind     = "aistore"
  basepath = "/arena-ml"
  endpoint = "http://10.1.1.2:51080"
  config = {
    auth = jsonencode({
      token = "top-secret"
    })
    max_objects = 1e6
    capacity_gb = 1000
  }
}
