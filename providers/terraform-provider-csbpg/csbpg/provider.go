// Package csbpg is a Terraform provider specialised for CSB PostgreSQL bindings
package csbpg

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"data_owner_role": {
				Type:     schema.TypeString,
				Required: true,
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

	host := d.Get("host").(string)
	port := d.Get("port").(int)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	database := d.Get("database").(string)
	dataOwnerRole := d.Get("data_owner_role").(string)

	return connectionFactory{
		host:          host,
		port:          port,
		username:      username,
		password:      password,
		database:      database,
		dataOwnerRole: dataOwnerRole,
	}, diags
}
