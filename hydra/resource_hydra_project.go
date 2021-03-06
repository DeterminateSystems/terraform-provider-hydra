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
			StateContext: schema.ImportStatePassthroughContext,
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

// Construct a PUT request to the /project/{id} endpoint that can either create
// a new project or update an existing one.
func createProjectPutBody(project string, d *schema.ResourceData) *api.PutProjectIdJSONRequestBody {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	homepage := d.Get("homepage").(string)
	owner := d.Get("owner").(string)

	body := api.PutProjectIdJSONRequestBody{
		Name:        &project,
		Displayname: &displayName,
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

	return &body
}

func resourceHydraProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to create project"
	client := m.(*api.ClientWithResponses)

	project := d.Get("name").(string)

	get, err := client.GetProjectIdWithResponse(ctx, project)
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
	body := createProjectPutBody(project, d)

	put, err := client.PutProjectIdWithResponse(ctx, project, *body)
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

	d.SetId(project)

	return nil
}

func resourceHydraProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to read Project"
	client := m.(*api.ClientWithResponses)

	id := d.Id()

	get, err := client.GetProjectIdWithResponse(ctx, id)
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

	projectResponse := get.JSON200

	d.Set("name", *projectResponse.Name)
	d.Set("display_name", *projectResponse.Displayname)
	d.Set("description", *projectResponse.Description)
	d.Set("homepage", *projectResponse.Homepage)
	d.Set("owner", *projectResponse.Owner)
	d.Set("enabled", *projectResponse.Enabled)
	d.Set("visible", !(*projectResponse.Hidden))

	// If Declarative can be dereferenced and every field is not empty, we update
	// the internal representation with it. If the project was never declarative,
	// none of these fields will be set. If it was ever declarative, some of these
	// fields may be set, and we want to inform Terraform of that.
	if projectResponse.Declarative != nil &&
		projectResponse.Declarative.File != nil &&
		projectResponse.Declarative.Type != nil &&
		projectResponse.Declarative.Value != nil &&
		!(*projectResponse.Declarative.File == "" &&
			*projectResponse.Declarative.Type == "" &&
			*projectResponse.Declarative.Value == "") {
		declarative := schema.NewSet(schema.HashResource(declInputSchema()), []interface{}{
			map[string]interface{}{
				"file":  *projectResponse.Declarative.File,
				"type":  *projectResponse.Declarative.Type,
				"value": *projectResponse.Declarative.Value,
			},
		})

		d.Set("declarative", declarative)
	} else {
		d.Set("declarative", nil)
	}

	d.SetId(*projectResponse.Name)

	return nil
}

func resourceHydraProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to update Project"
	client := m.(*api.ClientWithResponses)

	id := d.Id()
	newProject := d.Get("name").(string)
	body := createProjectPutBody(newProject, d)

	// Send the PUT request to the soon-to-be old project name using the resource's ID
	put, err := client.PutProjectIdWithResponse(ctx, id, *body)
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
		d.SetId(newProject)
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
