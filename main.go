package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/weesp-ai/terraform-provider-sendgrid/internal/provider"
)

// version is set by GoReleaser at build time via -ldflags.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/weesp-ai/sendgrid",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
