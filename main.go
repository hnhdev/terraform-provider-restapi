package main

import (
	"github.com/Mastercard/terraform-provider-restapi/restapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

// Generate SBOM
//go:generate go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod mod -json -output sbom.json

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return restapi.Provider()
		},
	})
}
