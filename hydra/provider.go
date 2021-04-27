package hydra

import (
	"context"
	"net/http"
	"net/http/cookiejar"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/publicsuffix"

	"terraform-provider-hydra/hydra/api"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HYDRA_HOST", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HYDRA_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HYDRA_PASSWORD", nil),
				Sensitive:   true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hydra_project": resourceHydraProject(),
			"hydra_jobset":  resourceHydraJobset(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	client := &http.Client{Jar: jar}

	c, err := api.NewClientWithResponses(host, func(c *api.Client) error {
		c.Client = client
		// Set JSON for all requests
		c.RequestEditors = append(c.RequestEditors,
			func(ctx context.Context, req *http.Request) error {
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")
				return nil
			})
		return nil
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	body := api.PostLoginJSONRequestBody{
		Username: &username,
		Password: &password,
	}
	resp, err := c.PostLoginWithResponse(ctx, body, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Origin", host)
		return nil
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer resp.HTTPResponse.Body.Close()

	if resp.JSON403 != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create Project",
			Detail:   *resp.JSON403.Error,
		})
	}

	return c, diags
}
