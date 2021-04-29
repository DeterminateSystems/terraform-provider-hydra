package hydra

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-hydra/hydra/api"
)

func nixExprSchema() *schema.Resource {
	return &schema.Resource{
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
	}
}

func inputSchema() *schema.Resource {
	return &schema.Resource{
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
						"boolean",
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
	}
}

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
				Default:     false,
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
				Elem:          nixExprSchema(),
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
				Elem:        inputSchema(),
			},
		},
	}
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

	// Check to make sure the project doesn't yet exist
	if getproj.HTTPResponse.StatusCode == http.StatusNotFound {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "Parent project does not exist.",
		}}
	}

	name := d.Get("name").(string)
	getjob, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, project, name)
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
	var state int
	var ty int

	switch d.Get("state").(string) {
	case "disabled":
		state = 0
	case "enabled":
		state = 1
	case "one-shot":
		state = 2
	case "one-at-a-time":
		state = 3
	}

	switch d.Get("type").(string) {
	case "legacy":
		ty = 0
	case "flake":
		ty = 1
	}

	visible := d.Get("visible").(bool)
	description := d.Get("description").(string)
	flake_uri := d.Get("flake_uri").(string)

	nix_expression := d.Get("nix_expression").(*schema.Set)

	check_interval := d.Get("check_interval").(int)
	scheduling_shares := d.Get("scheduling_shares").(int)

	email_notifications := d.Get("email_notifications").(bool)
	email_override := d.Get("email_override").(string)
	keep_evaluations := d.Get("keep_evaluations").(int)

	input := d.Get("input").(*schema.Set)

	if ty == 0 && flake_uri != "" {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify a flake_uri when using type \"legacy\".",
		}}
	}

	if ty == 1 && nix_expression != nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify a nix_expression when using type \"flake\".",
		}}
	}

	body := api.PutJobsetProjectIdJobsetIdJSONRequestBody{
		Project:          &project,
		Name:             &name,
		Enabled:          &state,
		Type:             &ty,
		Description:      &description,
		Checkinterval:    &check_interval,
		Schedulingshares: &scheduling_shares,
		Keepnr:           &keep_evaluations,
	}

	if visible {
		body.Visible = &visible
	}

	if email_notifications {
		body.Enableemail = &email_notifications
	}

	if email_override != "" {
		body.Emailoverride = &email_override
	}

	if flake_uri != "" {
		body.Flake = &flake_uri
	}

	if nix_expression != nil {
		// There will only ever be one nix_expression, so it's fine to access the
		// first (and only) element without precomputing
		expr := nix_expression.List()[0].(map[string]interface{})
		input := expr["in"].(string)
		path := expr["file"].(string)
		body.Nixexprinput = &input
		body.Nixexprpath = &path
	}

	if input != nil {
		inputs := make(map[string]api.JobsetInput)

		for _, value := range input.List() {
			v, _ := value.(map[string]interface{})
			name := v["name"].(string)
			ty := v["type"].(string)
			value := v["value"].(string)
			inputs[name] = api.JobsetInput{
				Name:  &name,
				Type:  &ty,
				Value: &value,
			}
		}

		body.Inputs = &api.Jobset_Inputs{
			AdditionalProperties: inputs,
		}
	}

	put, err := client.PutJobsetProjectIdJobsetIdWithResponse(ctx, project, name, body)
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

func resourceHydraJobsetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to read Jobset"
	client := m.(*api.ClientWithResponses)

	project := d.Get("project").(string)
	name := d.Id()

	get, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, project, name)
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
			Detail: fmt.Sprintf("Expected valid response from existing jobset, got %s:\n    %s",
				get.Status(), string(get.Body)),
		}}
	}

	jobset := get.JSON200

	var state string
	var ty string

	switch *jobset.Enabled {
	case 0:
		state = "disabled"
	case 1:
		state = "enabled"
	case 2:
		state = "one-shot"
	case 3:
		state = "one-at-a-time"
	}

	switch *jobset.Type {
	case 0:
		ty = "legacy"
	case 1:
		ty = "flake"
	}

	d.Set("project", *jobset.Project)
	d.Set("name", *jobset.Name)
	d.Set("state", state)
	d.Set("type", ty)
	d.Set("description", *jobset.Description)
	d.Set("check_interval", *jobset.Checkinterval)
	d.Set("scheduling_shares", *jobset.Schedulingshares)
	d.Set("keep_evaluations", *jobset.Keepnr)
	d.Set("visible", *jobset.Visible)
	d.Set("email_notifications", *jobset.Enableemail)

	if *jobset.Emailoverride != "" {
		d.Set("email_override", *jobset.Emailoverride)
	}

	if jobset.Nixexprinput != nil && jobset.Nixexprpath != nil {
		nix_expression := schema.NewSet(schema.HashResource(nixExprSchema()), []interface{}{
			map[string]interface{}{
				"in":   *jobset.Nixexprinput,
				"file": *jobset.Nixexprpath,
			},
		})

		d.Set("nix_expression", nix_expression)
	}

	if jobset.Inputs != nil {
		inputs := schema.NewSet(schema.HashResource(inputSchema()), flattenInputs(jobset.Inputs.AdditionalProperties))

		d.Set("input", inputs)
	}

	d.SetId(*jobset.Name)

	return nil
}

func resourceHydraJobsetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to update Project"
	client := m.(*api.ClientWithResponses)

	project := d.Get("project").(string)
	name := d.Get("name").(string)

	var state int
	var ty int

	switch d.Get("state").(string) {
	case "disabled":
		state = 0
	case "enabled":
		state = 1
	case "one-shot":
		state = 2
	case "one-at-a-time":
		state = 3
	}

	switch d.Get("type").(string) {
	case "legacy":
		ty = 0
	case "flake":
		ty = 1
	}

	visible := d.Get("visible").(bool)
	description := d.Get("description").(string)
	flake_uri := d.Get("flake_uri").(string)

	nix_expression := d.Get("nix_expression").(*schema.Set)

	check_interval := d.Get("check_interval").(int)
	scheduling_shares := d.Get("scheduling_shares").(int)

	email_notifications := d.Get("email_notifications").(bool)
	email_override := d.Get("email_override").(string)
	keep_evaluations := d.Get("keep_evaluations").(int)

	input := d.Get("input").(*schema.Set)

	if ty == 0 && flake_uri != "" {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify a flake_uri when using type \"legacy\".",
		}}
	}

	if ty == 1 && nix_expression != nil {
		return []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   "You cannot specify a nix_expression when using type \"flake\".",
		}}
	}

	body := api.PutJobsetProjectIdJobsetIdJSONRequestBody{
		Project:          &project,
		Name:             &name,
		Enabled:          &state,
		Type:             &ty,
		Description:      &description,
		Checkinterval:    &check_interval,
		Schedulingshares: &scheduling_shares,
		Keepnr:           &keep_evaluations,
	}

	if visible {
		body.Visible = &visible
	}

	if email_notifications {
		body.Enableemail = &email_notifications
	}

	if email_override != "" {
		body.Emailoverride = &email_override
	}

	if flake_uri != "" {
		body.Flake = &flake_uri
	}

	if nix_expression != nil {
		// There will only ever be one nix_expression, so it's fine to access the
		// first (and only) element without precomputing
		expr := nix_expression.List()[0].(map[string]interface{})
		input := expr["in"].(string)
		path := expr["file"].(string)
		body.Nixexprinput = &input
		body.Nixexprpath = &path
	}

	if input != nil {
		inputs := make(map[string]api.JobsetInput)

		for _, value := range input.List() {
			v, _ := value.(map[string]interface{})
			name := v["name"].(string)
			ty := v["type"].(string)
			value := v["value"].(string)
			inputs[name] = api.JobsetInput{
				Name:  &name,
				Type:  &ty,
				Value: &value,
			}
		}

		body.Inputs = &api.Jobset_Inputs{
			AdditionalProperties: inputs,
		}
	}

	put, err := client.PutJobsetProjectIdJobsetIdWithResponse(ctx, project, name, body)
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

	if d.HasChange("name") {
		d.SetId(name)
	}

	// Ensure we can still read the Jobset
	return resourceHydraJobsetRead(ctx, d, m)
}

func resourceHydraJobsetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	errsummary := "Failed to delete Jobset"
	client := m.(*api.ClientWithResponses)

	project := d.Get("project").(string)
	id := d.Id()

	del, err := client.DeleteJobsetProjectIdJobsetIdWithResponse(ctx, project, id)
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
	out := make([]interface{}, len(in), len(in))
	props := make(map[string]interface{})

	for k, v := range in {
		props["name"] = v.Name
		props["type"] = v.Type
		props["value"] = v.Value
		props["notify_committers"] = v.Emailresponsible
		out = append(out, props[k])
	}

	return out
}
