package hydra

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceHydraProject() *schema.Resource {
	return &schema.Resource{
		Description: "Resource defining a Hydra project.",

		CreateContext: resourceHydraProjectCreate,
		ReadContext:   resourceHydraProjectRead,
		UpdateContext: resourceHydraProjectUpdate,
		DeleteContext: resourceHydraProjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the project.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"display_name": {
				Description: "Display name of the project.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the project.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"homepage": {
				Description: "Homepage of the project.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"owner": {
				Description: "Owner of the project.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			// TODO: declarative configuration
			"enabled": {
				Description: "Whether or not the project is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"visible": {
				Description: "Whether or not the project is visible.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceHydraProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}

func resourceHydraProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}

func resourceHydraProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}

func resourceHydraProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}
