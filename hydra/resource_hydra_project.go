package hydra

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-hydra/hydra/api"
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
	client := m.(*api.ClientWithResponses)

	var diags diag.Diagnostics

	name := d.Get("name").(string)

	get, err := client.GetProjectIdWithResponse(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}
	defer get.HTTPResponse.Body.Close()

	// Check to make sure the project doesn't yet exist
	if get.HTTPResponse.StatusCode != http.StatusNotFound {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create Project",
			Detail:   "Project already exists.",
		})
	}

	// Now that we're sure the project doesn't exist, we can continue creating it
	display_name := d.Get("display_name").(string)
	description := d.Get("description").(string)
	homepage := d.Get("homepage").(string)
	owner := d.Get("owner").(string)
	enabled := d.Get("enabled").(bool)
	visible := d.Get("visible").(bool)

	body := api.PutProjectIdJSONRequestBody{
		Displayname: &display_name,
		Description: &description,
		Homepage:    &homepage,
		Owner:       &owner,
		Enabled:     &enabled,
		Visible:     &visible,
	}

	put, err := client.PutProjectIdWithResponse(ctx, name, body)
	if err != nil {
		return diag.FromErr(err)
	}
	defer put.HTTPResponse.Body.Close()

	// This should never happen (we login during the provider setup)
	if put.JSON403 != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create Project",
			Detail:   *put.JSON403.Error,
		})
	}

	if put.JSON201 == nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create Project",
			Detail:   "Expected successful project creation, got nil.",
		})
	}

	d.SetId(name)

	return diags
}

func resourceHydraProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*api.ClientWithResponses)

	var diags diag.Diagnostics

	name := d.Id()

	get, err := client.GetProjectIdWithResponse(ctx, name, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	defer get.HTTPResponse.Body.Close()

	// Check to make sure the project exists
	if get.HTTPResponse.StatusCode != http.StatusOK {
		d.SetId("")
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to read Project",
			Detail:   "Project does not exist.",
		})
	}

	project := get.JSON200

	d.Set("name", *project.Name)
	d.Set("display_name", *project.Displayname)
	d.Set("description", *project.Description)
	d.Set("homepage", *project.Homepage)
	d.Set("owner", *project.Owner)
	d.Set("enabled", *project.Enabled)
	d.Set("visible", !(*project.Hidden))
	d.SetId(*project.Name)

	return diags
}

func resourceHydraProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}

func resourceHydraProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented")
}
