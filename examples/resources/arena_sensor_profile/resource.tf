# resource "arenaml_sensor" "t1" {
#   name = ""
# }

# MinRateInHz: 6,
# MaxRateInHz: 13,
# MinRange:    30,
# MaxRange:    10000,
# RangeUnit:   "mm",
# Operating: map[string]float64{
# "Min Working Temp (C)": -10,
# "Max Working Temp (C)": 50,
# },
# Power: map[string]float64{
# "Min Input Voltage (V)":     4.5,
# "Max Input Voltage (V)":     5.5,
# "Typical Input Voltage (V)": 5.0,
# "Starting Current (mA)":     540,
# "Working Current (mA)":      290,
# },
# Comm: map[string]string{
# "Protocol":    "UART",
# "direction":   "one way",
# "Baud Rate":   "921600",
# "Data Length": "8bits",
# "Stop Bit":    "1",
# },
# Misc: map[string]interface{}{
# "Scanning Frequency":       "21600Hz",
# "Angular Resolution":       "0.167 deg",
# "Matching Life":            "10000 hours",
# "Dust and water resistant": "IP5X",
# },
# }

resource "arenaml_sensor_profile" "lidar_STL27L" {
  name        = "STL27L-LDRobot"
  kind        = "2d_liadar"
  description = "2d lidar with 360 deg coverage"
  spec = {
    h_fov          = 360
    v_fov          = 0.5
    range_unit     = "meter"
    max_range      = 25
    min_range      = 0.3
    max_rate_in_hz = 21600

    operating = {
      "Min Working Temp (C)" = -10
      "Max Working Temp (C)" = 50
    }

    power = {
      voltage                 = 5
      "Min Input Voltage"     = 4.5
      "Max Input Voltage"     = 5.5
      "Starting Current (mA)" = 540
      "Working Current (mA)"  = 290
    }

    comm = {
      "Protocol"    = "UART"
      "direction"   = "one way"
      "Baud Rate"   = "921600"
      "Data Length" = "8bits"
      "Stop Bit"    = "1"
    }

    misc = jsonencode({
      "Scanning Frequency"       = "21600Hz"
      "Angular Resolution"       = "0.167 deg"
      "Matching Life"            = "10000 hours"
      "Dust and water resistant" = "IP5X"
    })
  }
}


resource "arenaml_sensor_profile" "lidar_VL53L0X" {
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


