resource "arenaml_drone" "ugv_0" {
  name       = "ugv-0"
  kind       = "sbc"
  profile_id = "ugv-profile-id" // the spec values are then used from this drone profile
}

resource "arenaml_drone" "ugv_special" {
  name        = "ugv-rpi-5"
  kind        = "sbc"
  description = "ugv with rpi-5 as controller"
  spec = {
    arch         = "arm64"
    memory_in_gb = 8

    compute = {
      "4xA76" = 2.4,
      "total" = 9.6,
    }

    networks = {
      "wifi" = {
        kind      = "wifi-5"
        max_range = 10
        bandwidth = 50
      }
    }

    storage = {
      "sd-card" = {
        kind       = "micro-sd"
        capacity   = 118
        mount_path = "/"
        dev_path   = "/dev/mmcblk0p2"
      }
    }

    power = {
      capacity = "20000 mAh"
      output   = "5V 3A / 9V 2.3A"
    }

    details = {
      memory = "LPDDR4X",
    }
  }
}
