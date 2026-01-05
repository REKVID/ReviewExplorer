package models

type School struct {
	ID        int    `json:"id"`
	OrgType   string `json:"org_type"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
	Address   string `json:"address"`
	Website   string `json:"website"`
}
