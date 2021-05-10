package hydra

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-hydra/hydra/api"
)

func nixExprSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file": {
				Description: "The file in `input` which contains the Nix expression. Relative to the root of `input`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"input": {
				Description: "The name of the `input` which contains `file`.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func inputSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the input.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The type of the input.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "The value of the input.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"notify_committers": {
				Description: "Whether or not to notify committers.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceHydraJobset() *schema.Resource {
	return &schema.Resource{
		Description: "Resource defining a Hydra jobset.",

		CreateContext: resourceHydraJobsetCreate,
		ReadContext:   resourceHydraJobsetRead,
		UpdateContext: resourceHydraJobsetUpdate,
		DeleteContext: resourceHydraJobsetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"name": {
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
				Optional:    true,
				Default:     "Managed by terraform-provider-hydra.",
			},
			"flake_uri": {
				Description:   "(Mandatory when the `type` is `flake`, otherwise prohibited.) The jobset's flake URI.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"nix_expression"},
			},
			"nix_expression": {
				Description:   "(Mandatory when the `type` is `legacy`, otherwise prohibited.) The jobset's entrypoint Nix expression. The `file` must exist in an input which matches the name specified in `input`.",
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"flake_uri"},
				MinItems:      1,
				MaxItems:      1,
				Elem:          nixExprSchema(),
			},
			"check_interval": {
				Description: "How frequently to check the jobset in seconds (0 disables polling).",
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
				Optional:    true,
				Default:     false,
			},
			"email_override": {
				Description: "An email, or a comma-separated list of emails, to send email notifications to.",
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
				Elem:        inputSchema(),
			},
		},
	}
}

// Convert the state string specified in the resource config to an integer, as
// expected by the PUT API.
func stateToInt(state string) int {
	var s int

	switch state {
	case "disabled":
		s = 0
	case "enabled":
		s = 1
	case "one-shot":
		s = 2
	case "one-at-a-time":
		s = 3
	}

	return s
}

// Convert the GET response's state field (integer) to a string that would be
// valid in the resource config.
func stateToString(state int) string {
	var s string

	switch state {
	case 0:
		s = "disabled"
	case 1:
		s = "enabled"
	case 2:
		s = "one-shot"
	case 3:
		s = "one-at-a-time"
	}

	return s
}

// Convert the jobset type string specified in the resource config to an
// integer, as expected by the PUT API.
func jobsetTypeToInt(jobsetType string) int {
	var t int

	switch jobsetType {
	case "legacy":
		t = 0
	case "flake":
		t = 1
	}

	return t
}

// Convert the GET response's jobset type field (integer) to a string that would
// be valid in the resource config.
func jobsetTypeToString(jobsetType int) string {
	var t string

	switch jobsetType {
	case 0:
		t = "legacy"
	case 1:
		t = "flake"
	}

	return t
}

// Construct a PUT request to the /jobset/{project-id}/{jobset-id} endpoint that
// can either create a new jobset or update an existing one.
func createJobsetPutBody(project string, jobset string, d *schema.ResourceData) (*api.PutJobsetProjectIdJobsetIdJSONRequestBody, diag.Diagnostics) {
	errsummary := "Failed to create Jobset PUT request"

	// Collect errors so we can display all "syntax" errors that Terraform can't
	// handle for us at the same time (instead of one at a time).
	var diags diag.Diagnostics

	body := api.PutJobsetProjectIdJobsetIdJSONRequestBody{
		Project: &project,
		Name:    &jobset,
	}

	state := stateToInt(d.Get("state").(string))
	body.Enabled = &state

	jobsetType := jobsetTypeToInt(d.Get("type").(string))
	body.Type = &jobsetType

	description := d.Get("description").(string)
	body.Description = &description

	checkInterval := d.Get("check_interval").(int)
	body.Checkinterval = &checkInterval

	schedulingShares := d.Get("scheduling_shares").(int)
	body.Schedulingshares = &schedulingShares

	keepEvaluations := d.Get("keep_evaluations").(int)
	body.Keepnr = &keepEvaluations

	visible := d.Get("visible").(bool)
	if visible {
		body.Visible = &visible
	}

	emailNotifications := d.Get("email_notifications").(bool)
	if emailNotifications {
		body.Enableemail = &emailNotifications
	}

	emailOverride := d.Get("email_override").(string)
	if emailOverride != "" {
		body.Emailoverride = &emailOverride
	}

	flakeURI := d.Get("flake_uri").(string)
	if jobsetType == 0 && flakeURI != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify a flake_uri when using type \"legacy\".",
		})
	}

	if jobsetType == 1 && flakeURI == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Jobset type \"flake\" requires a non-empty flake_uri.",
		})
	}

	if flakeURI != "" {
		body.Flake = &flakeURI
	}

	nixExpression := d.Get("nix_expression").(*schema.Set)
	if jobsetType == 1 && len(nixExpression.List()) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify a nix_expression when using type \"flake\".",
		})
	}

	if jobsetType == 0 && len(nixExpression.List()) < 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Jobset type \"legacy\" requires a non-empty nix_expression.",
		})
	}

	if len(nixExpression.List()) > 0 {
		// There will only ever be one nix_expression, so it's fine to access the
		// first (and only) element without precomputing
		expr := nixExpression.List()[0].(map[string]interface{})
		input := expr["input"].(string)
		path := expr["file"].(string)
		body.Nixexprinput = &input
		body.Nixexprpath = &path
	}

	input := d.Get("input").(*schema.Set)
	if jobsetType == 1 && len(input.List()) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify one or more inputs when using type \"flake\".",
		})
	}

	if jobsetType == 0 && len(input.List()) < 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Jobset type \"legacy\" requires non-empty input(s).",
		})
	}

	if len(input.List()) > 0 {
		inputs := make(map[string]api.JobsetInput)

		for _, value := range input.List() {
			v, _ := value.(map[string]interface{})
			name := v["name"].(string)
			inputType := v["type"].(string)
			value := v["value"].(string)
			inputs[name] = api.JobsetInput{
				Name:  &name,
				Type:  &inputType,
				Value: &value,
			}
		}

		body.Inputs = &api.Jobset_Inputs{
			AdditionalProperties: inputs,
		}
	}

	return &body, diags
}

func resourceHydraJobsetParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected project/jobset", id)
	}

	return parts[0], parts[1], nil
}

func resourceHydraJobsetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to create jobset"
	client := m.(*api.ClientWithResponses)

	project := d.Get("project").(string)
	getproj, err := client.GetProjectIdWithResponse(ctx, project)
	if err != nil {
		return diag.FromErr(err)
	}
	defer getproj.HTTPResponse.Body.Close()

	// Check to make sure the parent project exists
	if getproj.HTTPResponse.StatusCode == http.StatusNotFound {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Parent project does not exist.",
		}}
	}

	jobset := d.Get("name").(string)
	getjob, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, project, jobset)
	if err != nil {
		return diag.FromErr(err)
	}
	defer getjob.HTTPResponse.Body.Close()

	// Check to make sure the jobset doesn't yet exist
	if getjob.HTTPResponse.StatusCode != http.StatusNotFound {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Jobset already exists.",
		}}
	}

	// Now that we're sure the jobset doesn't exist, we can continue creating it
	body, diags := createJobsetPutBody(project, jobset, d)
	if diags != nil {
		return diags
	}

	put, err := client.PutJobsetProjectIdJobsetIdWithResponse(ctx, project, jobset, *body)
	if err != nil {
		return diag.FromErr(err)
	}
	defer put.HTTPResponse.Body.Close()

	// If we didn't get the expected response, show what went wrong
	if put.JSON201 == nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail: fmt.Sprintf("Expected valid jobset creation response, got %s:\n    %s",
				put.Status(), string(put.Body)),
		}}
	}

	id := fmt.Sprintf("%s/%s", project, jobset)
	d.SetId(id)

	return nil
}

func resourceHydraJobsetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to read Jobset"
	client := m.(*api.ClientWithResponses)

	id := d.Id()

	project, jobset, err := resourceHydraJobsetParseID(id)
	if err != nil {
		return diag.FromErr(err)
	}

	get, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, project, jobset)
	if err != nil {
		return diag.FromErr(err)
	}
	defer get.HTTPResponse.Body.Close()

	// Check to make sure the jobset exists
	if get.HTTPResponse.StatusCode != http.StatusOK {
		d.SetId("")
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail: fmt.Sprintf("Expected valid response from existing jobset, got %s:\n    %s",
				get.Status(), string(get.Body)),
		}}
	}

	jobsetResponse := get.JSON200

	state := stateToString(*jobsetResponse.Enabled)
	jobsetType := jobsetTypeToString(*jobsetResponse.Type)

	d.Set("project", *jobsetResponse.Project)
	d.Set("name", *jobsetResponse.Name)
	d.Set("state", state)
	d.Set("type", jobsetType)
	d.Set("description", *jobsetResponse.Description)
	d.Set("check_interval", *jobsetResponse.Checkinterval)
	d.Set("scheduling_shares", *jobsetResponse.Schedulingshares)
	d.Set("keep_evaluations", *jobsetResponse.Keepnr)
	d.Set("visible", *jobsetResponse.Visible)
	d.Set("email_notifications", *jobsetResponse.Enableemail)

	if jobsetResponse.Emailoverride != nil && *jobsetResponse.Emailoverride != "" {
		d.Set("email_override", *jobsetResponse.Emailoverride)
	}

	if jobsetResponse.Flake != nil && *jobsetResponse.Flake != "" {
		d.Set("flake_uri", *jobsetResponse.Flake)
	}

	if (jobsetResponse.Nixexprinput != nil && *jobsetResponse.Nixexprinput != "") &&
		(jobsetResponse.Nixexprpath != nil && *jobsetResponse.Nixexprpath != "") {
		nixExpression := schema.NewSet(schema.HashResource(nixExprSchema()), []interface{}{
			map[string]interface{}{
				"input": *jobsetResponse.Nixexprinput,
				"file":  *jobsetResponse.Nixexprpath,
			},
		})

		d.Set("nix_expression", nixExpression)
	}

	if jobsetResponse.Inputs != nil {
		inputs := schema.NewSet(schema.HashResource(inputSchema()), flattenInputs(jobsetResponse.Inputs.AdditionalProperties))

		d.Set("input", inputs)
	}

	newID := fmt.Sprintf("%s/%s", *jobsetResponse.Project, *jobsetResponse.Name)
	d.SetId(newID)

	return nil
}

func resourceHydraJobsetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to update Jobset"
	client := m.(*api.ClientWithResponses)

	id := d.Id()
	newProject := d.Get("project").(string)
	newJobset := d.Get("name").(string)

	curProject, curJobset, err := resourceHydraJobsetParseID(id)
	if err != nil {
		return diag.FromErr(err)
	}

	body, diags := createJobsetPutBody(newProject, newJobset, d)
	if diags != nil {
		return diags
	}

	put, err := client.PutJobsetProjectIdJobsetIdWithResponse(ctx, curProject, curJobset, *body)
	if err != nil {
		return diag.FromErr(err)
	}
	defer put.HTTPResponse.Body.Close()

	// If we didn't get the expected response, show what went wrong
	if put.JSON200 == nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail: fmt.Sprintf("Expected valid reponse from existing jobset, got %s:\n    %s",
				put.Status(), string(put.Body)),
		}}
	}

	if d.HasChange("name") || d.HasChange("project") {
		id := fmt.Sprintf("%s/%s", newProject, newJobset)
		d.SetId(id)
	}

	// Ensure we can still read the Jobset
	return resourceHydraJobsetRead(ctx, d, m)
}

func resourceHydraJobsetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to delete Jobset"
	client := m.(*api.ClientWithResponses)

	id := d.Id()

	project, jobset, err := resourceHydraJobsetParseID(id)
	if err != nil {
		return diag.FromErr(err)
	}

	del, err := client.DeleteJobsetProjectIdJobsetIdWithResponse(ctx, project, jobset)
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

func flattenInputs(in map[string]api.JobsetInput) []interface{} {
	out := make([]interface{}, 0, len(in))

	for k := range in {
		props := make(map[string]interface{})

		props["name"] = *in[k].Name
		props["notify_committers"] = *in[k].Emailresponsible
		props["type"] = *in[k].Type
		props["value"] = *in[k].Value
		out = append(out, props)
	}

	return out
}
