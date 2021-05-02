package hydra

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-hydra/hydra/api"
)

func declInputSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file": {
				Description: "The file in `value` which contains the declarative spec file. Relative to the root of `input`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The type of the declarative input.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "The value of the declarative input.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceHydraProject() *schema.Resource {
	return &schema.Resource{
		Description: "Resource defining a Hydra project.",

		CreateContext: resourceHydraProjectCreate,
		ReadContext:   resourceHydraProjectRead,
		UpdateContext: resourceHydraProjectUpdate,
		DeleteContext: resourceHydraProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Optional:    true,
				Default:     "Managed by terraform-provider-hydra.",
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
			"declarative": {
				Description: "Configuration of the declarative project.",
				Type:        schema.TypeSet,
				Optional:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem:        declInputSchema(),
			},
		},
	}
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

	body := api.PutProjectIdJSONRequestBody{
		Name:        &name,
		Displayname: &display_name,
		Description: &description,
		Homepage:    &homepage,
		Owner:       &owner,
	}

	enabled := d.Get("enabled").(bool)
	if enabled {
		body.Enabled = &enabled
	}

	visible := d.Get("visible").(bool)
	if visible {
		body.Visible = &visible
	}

	declarative := d.Get("declarative").(*schema.Set)
	if len(declarative.List()) > 0 {
		// There will only ever be one declarative block, so it's fine to access the
		// first (and only) element without precomputing
		decl := declarative.List()[0].(map[string]interface{})
		file := decl["file"].(string)
		inputType := decl["type"].(string)
		value := decl["value"].(string)

		body.Declarative = &api.DeclarativeInput{
			File:  &file,
			Type:  &inputType,
			Value: &value,
		}
	}

	put, err := client.PutProjectIdWithResponse(ctx, name, body)
	if err != nil {
		return diag.FromErr(err)
	}
	defer put.HTTPResponse.Body.Close()

	// If we didn't get the expected response, show what went wrong
	if put.JSON201 == nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail: fmt.Sprintf("Expected valid project creation response, got %s:\n    %s",
				put.Status(), string(put.Body)),
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
			Detail: fmt.Sprintf("Expected valid response from existing project, got %s:\n    %s",
				get.Status(), string(get.Body)),
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

	if project.Declarative != nil &&
		(project.Declarative.File != nil && *project.Declarative.File != "") &&
		(project.Declarative.Type != nil && *project.Declarative.Type != "") &&
		(project.Declarative.Value != nil && *project.Declarative.Value != "") {
		declarative := schema.NewSet(schema.HashResource(declInputSchema()), []interface{}{
			map[string]interface{}{
				"file":  *project.Declarative.File,
				"type":  *project.Declarative.Type,
				"value": *project.Declarative.Value,
			},
		})

		d.Set("declarative", declarative)
	}

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
	declarative := d.Get("declarative").(*schema.Set)

	body := api.PutProjectIdJSONRequestBody{
		Name:        &name,
		Displayname: &display_name,
		Description: &description,
		Homepage:    &homepage,
		Owner:       &owner,
	}

	if enabled {
		body.Enabled = &enabled
	}

	if visible {
		body.Visible = &visible
	}

	if len(declarative.List()) > 0 {
		// There will only ever be one declarative block, so it's fine to access the
		// first (and only) element without precomputing
		decl := declarative.List()[0].(map[string]interface{})
		file := decl["file"].(string)
		inputType := decl["type"].(string)
		value := decl["value"].(string)

		body.Declfile = &file
		body.Decltype = &inputType
		body.Declvalue = &value
	}

	// Send the PUT request to the soon-to-be old project name using the resource's ID
	put, err := client.PutProjectIdWithResponse(ctx, id, body)
	if err != nil {
		return diag.FromErr(err)
	}
	defer put.HTTPResponse.Body.Close()

	// If we didn't get the expected response, show what went wrong
	if put.JSON200 == nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail: fmt.Sprintf("Expected valid response from existing project, got %s:\n    %s",
				put.Status(), string(put.Body)),
		}}
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
			Detail: fmt.Sprintf("Expected valid project deletion response, got %s:\n    %s",
				del.Status(), string(del.Body)),
		}}
	}

	d.SetId("")

	return nil
}
