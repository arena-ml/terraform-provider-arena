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
