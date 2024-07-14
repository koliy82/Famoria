package user

type User struct {
	ID                    uint64 `ch:"id"`
	IsBot                 bool   `ch:"is_bot"`
	FirstName             string `ch:"first_name"`
	LastName              string `ch:"last_name"`
	Username              string `ch:"username"`
	LanguageCode          string `ch:"language_code"`
	IsPremium             bool   `ch:"is_premium"`
	AddedToAttachmentMenu bool   `ch:"added_to_attachment_menu"`
	IsAdmin               bool   `ch:"is_admin"`
}
