package model

type User struct {
	ID            string   `json:"id" bson:"_id,omitempty"`
	Email         string   `json:"email"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	Authenticated bool     `json:"authenticated"`
	ProjectIDs    []string `json:"projectIds"`
}

type Project struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"ownerId"`
}

func (p *Project) IsOwner(user *User) bool {
	return p.OwnerID == user.ID
}
