package hydra

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/publicsuffix"

	"terraform-provider-hydra/hydra/api"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Description: "The address of the Hydra instance.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HYDRA_HOST", nil),
			},
			"username": {
				Description: "The user that Terraform will be logging in as.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HYDRA_USERNAME", nil),
			},
			"password": {
				Description: "The password for the Hydra user specified in `username`.",
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
		ConfigureContextFunc: providerConfigure,
	}
}

// RetryableHTTPClient is a wrapper type around retryablehttp.Client so that we
// can implement the HttpRequestDoer interface.
type RetryableHTTPClient retryablehttp.Client

// Do - Perform the provided HTTP request using a retryablehttp.Client.
func (c *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}

	client := (*retryablehttp.Client)(c)

	return client.Do(r)
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	errsummary := "Failed to configure Provider"

	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	retry := retryablehttp.NewClient()
	retry.RetryWaitMin = time.Second
	retry.RetryWaitMax = 30 * time.Second
	retry.RetryMax = 10
	retry.CheckRetry = retryPolicy
	retry.Logger = nil
	retry.HTTPClient.Jar = jar
	retry.HTTPClient.Transport = logging.NewTransport(
		"DeterminateSystems/terraform-provider-hydra",
		retry.HTTPClient.Transport,
	)

	httpclient := (*RetryableHTTPClient)(retry)

	client, err := api.NewClientWithResponses(host, func(c *api.Client) error {
		c.Client = httpclient
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

	resp, err := client.PostLoginWithResponse(ctx, body,
		func(ctx context.Context, req *http.Request) error {
			origin, err := url.Parse(host)
			if err != nil {
				return err
			}

			// Set the User field to nil in order to strip out authentication information
			// from the URI -- Hydra expects *only* the host (and port, if necessary) as
			// the Origin.
			origin.User = nil

			req.Header.Add("Origin", origin.String())
			return nil
		})
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer resp.HTTPResponse.Body.Close()

	if resp.JSON403 != nil {
		return nil, []diag.Diagnostic{{
			Severity: diag.Error,
			Summary:  errsummary,
			Detail:   *resp.JSON403.Error,
		}}
	}

	return client, nil
}

// https://github.com/packethost/terraform-provider-packet/blob/c57d85cfe55288a87b51938ff8909fdbf932a5af/packet/config.go#L24
var redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, nil
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, nil
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}

	return false, nil
}
