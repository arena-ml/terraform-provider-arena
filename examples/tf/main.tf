# Copyright (c) ArenaML Labs Pvt Ltd.

terraform {
  required_providers {
    arenaml = {
      source  = "arena-io/arena"
      version = "0.0.8"
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
