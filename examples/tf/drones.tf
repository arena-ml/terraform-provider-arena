
resource "arenaml_drone_profile" "rx" {
  description = "2d lidar"
  kind        = "LIDAR_2D"
  name        = "dfrobota"
  spec = {
    compute = {
      "gpu"  = 2
      "cuda" = 13
    }
    storage = {
      main = {
        guid     = "1234"
        dev_path = "/mnt/dev"
        capacity = 9
      }
    }
  }
}
