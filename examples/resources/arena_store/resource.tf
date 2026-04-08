
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