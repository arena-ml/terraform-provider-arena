# locally running nomad cluster


// adding a local server for demo with `nomad agent -dev`
resource "arenaml_cluster_manager" "local" {
  name = "eg-nomad-local"
  kind = "nomad"
  spec = jsonencode({
    address = "http://loaclhost:4646"
  })
}

resource "arenaml_cluster_manager" "pass_protected" {
  name = "eg-nomad-http-auth"
  kind = "nomad"
  spec = jsonencode({
    address = "https://my.nomad.server"
    http_auth = {
      username = "such_user"
      password = "much_password"
    }
  })
}

resource "arenaml_cluster_manager" "tls_certs" {
  name = "eg-nomad-tls"
  kind = "nomad"
  spec = jsonencode({
    address = "https://safe.nomad.server"
    tls_config = {
      client_cert = "/very/safe/file.crt"
      client_key  = "/very/safe/file.pem"
      ca_cert     = "/very/safe/file.ca"
    }
  })
}




# for nomad spec, the field types and json key as per the below schema. more details can be found here : https://github.com/hashicorp/nomad/blob/d9d98c0866288a23722dbf31637640d02bd4fb6f/api/api.go#L171

# Address   string `json:"address" yaml:"address"`
# Region    string `json:"region" yaml:"region"`
# ID        string `json:"id" yaml:"id"`
# SecretID  string `json:"secret_id" yaml:"secret_id"`
# Namespace string `json:"namespace" yaml:"namespace"`
# HttpAuth  HttpBasicAuth `json:"http_auth"`
# TLSConfig TLSConfig `json:"tls_config"`

# -----

// HttpBasicAuth is used to authenticate http client with HTTP Basic Authentication
# type HttpBasicAuth struct {

// Username to use for HTTP Basic Authentication
# Username string  `json:"username"`

// Password to use for HTTP Basic Authentication
# Password string  `json:"password"`

# -----
# type TLSConfig struct {

# CACert string `json:"ca_cert"`
# CAPath string `json:"ca_path"`
# CACertPEM []byte `json:"ca_cert_pem"`
# ClientCert string `json:"client_cert"`
# ClientCertPEM []byte `json:"client_cert_pem"`
# ClientKey string `json:"client_key"`
# ClientKeyPEM []byte `json:"client_key_pem"`
# TLSServerName string `json:"tls_server_name"`
# Insecure bool `json:"insecure"`

# }


