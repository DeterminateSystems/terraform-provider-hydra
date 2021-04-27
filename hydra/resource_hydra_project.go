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

func checkPutProjectId(put *api.PutProjectIdResponse, summary string) diag.Diagnostics {
	// This should never happen (we login during the provider setup)
	if put.JSON403 != nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   *put.JSON403.Error,
		}}
	}

	if put.JSON400 != nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   *put.JSON400.Error,
		}}
	}

	return nil
}

func resourceHydraProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to create project"
	client := m.(*api.ClientWithResponses)

	name := d.Get("name").(string)

	get, err := client.GetProjectIdWithResponse(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}
	defer get.HTTPResponse.Body.Close()

	// Check to make sure the project doesn't yet exist
	if get.HTTPResponse.StatusCode != http.StatusNotFound {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Project already exists.",
		}}
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

	if err := checkPutProjectId(put, errsummary); err != nil {
		return err
	}

	if put.JSON201 == nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Expected project response, got nil.",
		}}
	}

	d.SetId(name)

	return nil
}

func resourceHydraProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to read Project"
	client := m.(*api.ClientWithResponses)

	name := d.Id()

	get, err := client.GetProjectIdWithResponse(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}
	defer get.HTTPResponse.Body.Close()

	// Check to make sure the project exists
	if get.HTTPResponse.StatusCode != http.StatusOK {
		d.SetId("")
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Project does not exist.",
		}}
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

	return nil
}

func resourceHydraProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to update Project"
	client := m.(*api.ClientWithResponses)

	id := d.Id()
	name := d.Get("name").(string)
	display_name := d.Get("display_name").(string)
	description := d.Get("description").(string)
	homepage := d.Get("homepage").(string)
	owner := d.Get("owner").(string)
	enabled := d.Get("enabled").(bool)
	visible := d.Get("visible").(bool)

	body := api.PutProjectIdJSONRequestBody{
		Name:        &name,
		Displayname: &display_name,
		Description: &description,
		Homepage:    &homepage,
		Owner:       &owner,
		Enabled:     &enabled,
		Visible:     &visible,
	}

	// Send the PUT request to the soon-to-be old project name using the resource's ID
	put, err := client.PutProjectIdWithResponse(ctx, id, body)
	if err != nil {
		return diag.FromErr(err)
	}
	defer put.HTTPResponse.Body.Close()

	if err := checkPutProjectId(put, errsummary); err != nil {
		return err
	}

	if d.HasChange("name") {
		d.SetId(name)
	}

	// Ensure we can still read the Project
	return resourceHydraProjectRead(ctx, d, m)
}

func resourceHydraProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to delete Project"
	client := m.(*api.ClientWithResponses)

	id := d.Id()

	del, err := client.DeleteProjectIdWithResponse(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	defer del.HTTPResponse.Body.Close()

	// Check to make sure the project was actually deleted
	if del.HTTPResponse.StatusCode != http.StatusOK {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Project does not exist.",
		}}
	}

	d.SetId("")

	return nil
}
