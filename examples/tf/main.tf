# Copyright (c) ArenaML Labs Pvt Ltd.

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

# data "arenaml_engine" "default" {
#   id = "00000000-0000-4000-9000-000000000000"
# }
#
# data "arenaml_engine" "old" {
#   id = "00000000-0000-4000-9000-000000000000"
# }
#
# output "default_engine" {
#   value = data.arenaml_engine.default
# }
#
# output "old_engine" {
#   value = data.arenaml_engine.old
# }
#
resource "arenaml_cluster_manager" "def" {
  name = "tf_test_engine"
  spec = file("${path.module}/default-engine-spec.json")
}
#
#
#
# data "arenaml_basis" "some_basis" {
#   id = "d4eb5b0f-215e-4c63-8430-675f00e9f529"
# }
#
# output "sb" {
#   value = data.arenaml_basis.some_basis
# }
#
# resource "arenaml_basis" "test_basis_2" {
#   description = "biggram training and prompt"
#   kind        = "git"
#   name        = "tf_test_2"
#   source = {
#     format = "toml"
#     raw    = <<-EOT
#
#     uri="https://github.com/arena-ml/bigram_train_sp.git"
#
#     [collect]
#     max_new_versions=1
#
#     EOT
#   }
#   watcher = {
#     image = "reg.arenaml.dev/basis-contrib/git:latest"
#   }
# }
#
# data "arenaml_pipeline_input" "some_input" {
#   id = "bd045ac6-6e8c-4e84-9f54-1562b30b1994"
# }
#
# output "si" {
#   value = data.arenaml_pipeline_input.some_input
# }
#
# resource "arenaml_pipeline_input" "test_input" {
#   pipeline_id = "00000000-0000-4000-a000-000000000000"
#   name = "foo-bar"
# }
#
# resource "arenaml_pipeline" "some_pkl" {
#   name        = "some-pl"
#   org_id      = "5ca1ab1e-0000-4000-a000-000000000000"
#   description = "some pipeline"
# }
#
# resource "arenaml_pipeline_input" "input_a" {
#   name        = "in-A"
#   pipeline_id = arenaml_pipeline.some_pkl.id
# }
#
#
# resource "arenaml_pipeline_step" "step_a" {
#   pipeline_id = arenaml_pipeline.some_pkl.id
#   name        = "step-a"
#   kind        = "docker"
#   config = {
#     image = "debian:latest"
#   }
# }
#
# resource "arenaml_pipeline_output" "out_b" {
#   name        = "out-a"
#   pipeline_id = arenaml_pipeline.some_pkl.id
#
# }
#
#
# resource "arenaml_pipeline_dag" "okl_dag" {
#   pipeline_id = arenaml_pipeline.some_pkl.id
#
#   input_edges = [
#     {
#       node_id    = arenaml_pipeline_input.input_a.id
#       from_bases = [arenaml_basis.test_basis_2.id]
#       to_steps   = [arenaml_pipeline_step.step_a.id]
#       # from_inputs = []
#       # to_inputs = []
#     }
#   ]
#
#   output_edges = [
#     {
#       node_id   = arenaml_pipeline_output.out_b.id
#       from_step = arenaml_pipeline_step.step_a.id
#       # to_inputs = []
#     }
#   ]
# }

# data "arenaml_org" "some_org" {
#   id = "5ca1ab1e-0000-4000-a000-000000000000"
# }
#
data "arenaml_user" "some_user" {
  id = "00000000-0000-4000-9000-000000000000"
}
#
# data "arenaml_team" "some_team" {
#     id = "decade00-0000-4000-a000-000000000000"
# }
#
# data "arenaml_store" "some_store" {
#   id = "00000000-0000-4000-8000-000000000000"
# }

#
# data "arenaml_sensor_profile" "a" {
#   id = "e486bf3a-4705-4ef6-948b-7beeaf1371e5"
# }


data "arenaml_drone" "a" {
  id = "deb00000-0000-4000-a000-000000000000"
}
#
# data "arenaml_drone_profile" "a" {
#   id = "deb00000-0000-4000-8000-000000000000"
# }

#
# data "arena_drone_profile" "a" {
#   id = "deb00000-0000-4000-a000-000000000000"
# }
#
output "sensor_A" {
  value = data.arenaml_drone.a
}

#
# output "sensor_pro_A" {
#   value = arenaml_pipeline_input.input_a
# }


# output "some_org" {
#   value = data.arenaml_org.some_org
# }
#
# output "some_team" {
#   value = data.arenaml_team.some_team
# }
#
# output "some_user" {
#   value = data.arenaml_store.some_store
#   sensitive = true
# }
#
#
#
#
#
#
#
