package hydra

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHydraJobset() *schema.Resource {
	return &schema.Resource{
		Description: "Resource defining a Hydra jobset.",

		CreateContext: resourceHydraJobsetCreate,
		ReadContext:   resourceHydraJobsetRead,
		UpdateContext: resourceHydraJobsetUpdate,
		DeleteContext: resourceHydraJobsetDelete,

		Schema: map[string]*schema.Schema{
			"project": {
				Description: "Name of the parent project.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"state": {
				Description: "State of the jobset.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"enabled",
					"one-shot",
					"one-at-a-time",
					"disabled",
				}, false),
			},
			"visible": {
				Description: "Whether or not the jobset is visible.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"name": { // Identifier
				Description: "Name of the jobset.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "Type of jobset.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"legacy", "flake"}, false),
			},
			"description": {
				Description: "Description of the jobset.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"flake_uri": {
				Description:   "(Mandatory when the `type` is `flake`, otherwise prohibited.) The jobset's flake URI.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"nix_expression"},
			},
			"nix_expression": {
				Description:   "(Mandatory when the `type` is `legacy`, otherwise prohibited.) The jobset's entrypoint Nix expression. The `file` must exist in an input, matching the `in` name.",
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"flake_uri"},
				MaxItems:      1,
				MinItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file": {
							Type:     schema.TypeString,
							Required: true,
						},
						"in": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"check_interval": {
				Description: "How frequently to check the jobset in seconds.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"scheduling_shares": {
				Description: "How many shares allocated to the jobset.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"email_notifications": {
				Description: "Whether or not to send email notifications",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"email_override": {
				Description: "Where to send email notifications.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"keep_evaluations": {
				Description: "How many of the jobset's evaluations to keep.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"input": {
				Description: "Input(s) provided to the jobset.",
				Type:        schema.TypeSet,
				Optional:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"bool",
									"bitbucketpulls",
									"git",
									"darcs",
									"path",
									"githubpulls",
									"svn",
									"svn-checkout",
									"bzr",
									"bzr-checkout",
									"gitlabpulls",
									"hg",
									"github_refs",
								}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"notify_committers": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceHydraJobsetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: type flake and nix_expression are mutually exclusive
	// TODO: type legacy and flake_uri are mutually exclusive
	return diag.Errorf("not implemented")
}

func resourceHydraJobsetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}

func resourceHydraJobsetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}

func resourceHydraJobsetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}
