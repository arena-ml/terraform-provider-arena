package main

import (
	"context"
	"flag"
	"log"

	"github.com/arena-ml/terraform-provider-arenaml/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go tool github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir . -provider-name arena

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "arenaml.dev/tf/arenaml",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
