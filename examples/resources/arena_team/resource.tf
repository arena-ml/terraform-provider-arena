resource "arenaml_org" "nav_dev" {
  name        = "nav-dev"
  description = "developer group for navigation systems"
}

resource "arenaml_team" "auto_flight" {
  name        = "path-planning"
  role        = "devs"
  org_id      = arena_org.nav_dev.id // teams must have a parent org
  description = "dev team working on autonomous path planning"
  config = {
    allow_cross_orgs = true
  }
}
