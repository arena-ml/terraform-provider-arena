# resource "arenaml_sensor" "t1" {
#   name = ""
# }

# Name:        "dfrobot",
# Description: "2d lidar",
# Kind:        "LIDAR_2D",
# Spec: &schema.SensorSpec{
# MinRateInHz: 10,
# MinRange:    1,
# MaxRange:    100000000,
# Operating: map[string]float64{
# "Min Opertating Temp (C)": -30,
# "Max Opertating Temp (C)": 85,
# },


resource "arenaml_sensor_profile" "a" {
  name = "STL27L-LDRobot"
  kind = "2d_lidar"
}