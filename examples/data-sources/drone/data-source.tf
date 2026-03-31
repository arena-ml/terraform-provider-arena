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

resource "arena_drone" "mav_13" {
  name = "mav_13"
}