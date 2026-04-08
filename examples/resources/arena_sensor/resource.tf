resource "arenaml_sensor" "lidar_2D" {
  name        = "STL27L-LDRobot"
  kind        = "2d_liadar"
  description = "2d lidar with 360 deg coverage"
  profile_id  = "STL27L_profile_id" // spec values are taken from the sensor profile
}


resource "arenaml_sensor" "range_lidar" {
  name        = "VL53L0X"
  kind        = "point_liadar"
  description = "range finding point lidar"
  spec = {
    h_fov          = 25
    v_fov          = 25
    range_unit     = "meter"
    max_range      = 2
    min_range      = 0.05
    max_rate_in_hz = 20

    power = {
      voltage                = 5
      "Min Input Voltage"    = 2.6
      "Max Input Voltage"    = 3.5
      "Peak Current (mA)"    = 40
      "Working Current (mA)" = 10
    }

    comm = {
      "Protocol"            = "I2C"
      "Operating frequency" = "400 KHz"
    }

    operating = {
      "Min Working Temp (C)" = -20
      "Max Working Temp (C)" = 70
    }
  }
}


