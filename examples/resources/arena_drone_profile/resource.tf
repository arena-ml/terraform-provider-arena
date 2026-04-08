resource "arenaml_drone_profile" "rpi_5" {
  name        = "rpi-5"
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

resource "arenaml_drone_profile" "radxa_5T" {
  name        = "radxa-5T"
  kind        = "sbc"
  description = "ugv with radxa 5T as controller"
  spec = {
    arch         = "arm64"
    memory_in_gb = 24

    compute = {
      "gpu"  = 2
      "cuda" = 13
    }

    npu = {
      tops = 6
    }

    networks = {
      "wifi" = {
        kind      = "wifi-6"
        max_range = 16
        bandwidth = 100
      }
    }


    storage = {
      main = {
        capacity   = 931
        kind       = "nvme"
        dev_path   = "/dev/nvme0n1p3"
        mount_path = "/"
      }
    }

    power = {
      capacity = "10400 mAh"
      output   = "12V 3A"
    }


    details = {
      memory      = "LPDDR5",
      SoC         = "Broadcom BCM2712",
      gpu         = "Arm Mali G610MC4"
      opengl      = "ES1.1, ES2.0, and ES3.2"
      opencl      = "1.1, 1.2 and 2.2"
      nvme        = "M.2 M with PCIe 3.0 2-lane"
      power_input = " 12V 5525 DC Jack"
      dimension   = "110 mm x 82 mm"
    }
  }
}