# Copyright (c) ArenaML Labs Pvt Ltd.

terraform {
  required_providers {
    arena = {
      source  = "arena-io/arena"
      version = "0.0.8"
    }
  }
}

provider "arena" {
  server_url = "http://localhost:18080/api/v1"
}


resource "arena_cluster_manager" "def" {
  name = "tf_test_engine"
  kind = "nomad"
  spec = file("${path.module}/default-engine-spec.json")
}
