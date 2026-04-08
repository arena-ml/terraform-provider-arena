data "arenaml_basis" "model_repo" {
  id = "b14bf594-52e6-4aa8-92b3-99bc6389069c"
}

resource "arenaml_basis" "test_basis" {
  description = "bi-gram training and prompt"
  kind        = "git"
  name        = "tf_test"

  source = {
    format = "toml"
    raw    = <<-EOT

     uri="https://github.com/arena-ml/bigram_train_sp.git"

     [collect]
     max_new_versions=1

     EOT
  }

  watcher = {
    image      = "reg.arenaml.dev/basis-contrib/git:latest"
    no_collect = true
  }
}
