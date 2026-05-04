package user_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/view"
)

// UserListItemView wraps User for template rendering with CSP nonce
type UserListItemView struct {
	view.BaseTemplateData
	User    entity_public.User
	IsAdmin bool
}
