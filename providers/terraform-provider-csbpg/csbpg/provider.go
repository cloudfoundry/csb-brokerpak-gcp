// Package csbpg is a Terraform provider specialised for CSB PostgreSQL bindings
package csbpg

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	dataOwnerRoleKey = "data_owner_role"
	databaseKey      = "database"
	passwordKey      = "password"
	usernameKey      = "username"
	portKey          = "port"
	hostKey          = "host"
	sslModeKey       = "sslmode"
	clientCertKey    = "clientcert"
	sslRootCertKey   = "sslrootcert"
)

func Provider() *schema.Provider {

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			hostKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			portKey: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			usernameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			passwordKey: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			databaseKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			dataOwnerRoleKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			sslModeKey: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "verify-ca",
				Description: "This option determines whether or with what priority a secure SSL TCP/IP connection will be negotiated with the PostgreSQL server",
			},
			clientCertKey: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "SSL client certificate if required by the database.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cert": {
							Type:        schema.TypeString,
							Description: "The SSL client certificate file path, must contain PEM encoded data.",
							Required:    true,
						},
						"key": {
							Type:        schema.TypeString,
							Description: "The SSL client certificate private key, must contain PEM encoded data.",
							Required:    true,
						},
					},
				},
				MaxItems: 1,
			},
			sslRootCertKey: {
				Type:        schema.TypeString,
				Description: "The SSL server root, must contain PEM encoded data.",
				Optional:    true,
			},
		},
		ConfigureContextFunc: providerConfigure,
		ResourcesMap: map[string]*schema.Resource{
			"csbpg_binding_user": resourceBindingUser(),
		},
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	factory := connectionFactory{
		host:          d.Get(hostKey).(string),
		port:          d.Get(portKey).(int),
		username:      d.Get(usernameKey).(string),
		password:      d.Get(passwordKey).(string),
		database:      d.Get(databaseKey).(string),
		dataOwnerRole: d.Get(dataOwnerRoleKey).(string),
		sslMode:       d.Get(sslModeKey).(string),
		sslRootCert:   d.Get(sslRootCertKey).(string),
	}

	if value, ok := d.GetOk(clientCertKey); ok {
		if spec, ok := value.([]interface{})[0].(map[string]interface{}); ok {
			factory.sslClientCert = &clientCertificateConfig{
				Certificate: spec["cert"].(string),
				Key:         spec["key"].(string),
			}
		}
	}

	return factory, diags
}
