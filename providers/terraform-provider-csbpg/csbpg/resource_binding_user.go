package csbpg

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lib/pq"
)

const (
	bindingUsernameKey = "username"
	bindingPasswordKey = "password"
)

func resourceBindingUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			bindingUsernameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			bindingPasswordKey: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
		CreateContext: resourceBindingUserCreate,
		ReadContext:   resourceBindingUserRead,
		UpdateContext: resourceBindingUserUpdate,
		DeleteContext: resourceBindingUserDelete,
		Description:   "TODO",
		UseJSONNumber: true,
	}
}

func resourceBindingUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {

	log.Println("[DEBUG] ENTRY resourceBindingUserCreate()")
	defer log.Println("[DEBUG] EXIT resourceBindingUserCreate()")

	username := d.Get(bindingUsernameKey).(string)
	password := d.Get(bindingPasswordKey).(string)

	id := fmt.Sprintf("bindinguser/%s", username)

	cf := m.(connectionFactory)

	db, err := cf.Connect()
	if err != nil {
		return diag.FromErr(err)
	}
	defer db.Close()
	log.Println("[DEBUG] connected")

	err = createDataOwnerRole(db, cf)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("[DEBUG] create binding user")
	_, err = db.Exec(fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD %s INHERIT IN ROLE %s", pq.QuoteIdentifier(username), safeQuote(password), pq.QuoteIdentifier(cf.dataOwnerRole)))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] setting ID %s\n", id)
	d.SetId(id)

	return nil
}

func resourceBindingUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return nil
}

func resourceBindingUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return nil
}

func resourceBindingUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {

	log.Println("[DEBUG] ENTRY resourceBindingUserDelete()")
	defer log.Println("[DEBUG] EXIT resourceBindingUserDelete()")

	username := d.Get(bindingUsernameKey).(string)

	cf := m.(connectionFactory)

	db, err := cf.Connect()
	if err != nil {
		return diag.FromErr(err)
	}
	defer db.Close()
	log.Println("[DEBUG] connected")

	log.Println("[DEBUG] dropping binding user")
	_, err = db.Exec(fmt.Sprintf("DROP ROLE %s", pq.QuoteIdentifier(username)))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func safeQuote(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `'`, `\\`))
}

func roleExists(db *sql.DB, name string) (bool, error) {
	log.Println("[DEBUG] ENTRY roleExists()")
	defer log.Println("[DEBUG] EXIT roleExists()")

	rows, err := db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", name))
	if err != nil {
		return false, fmt.Errorf("error finding role %q: %w", name, err)
	}
	defer rows.Close()
	return rows.Next(), nil
}
