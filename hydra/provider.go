package hydra

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"hydra_project": resourceHydraProject(),
			"hydra_jobset":  resourceHydraJobset(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
}
