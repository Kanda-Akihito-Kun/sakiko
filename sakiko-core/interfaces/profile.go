package interfaces

type Profile struct {
	ID         string      `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	Source     string      `json:"source" yaml:"source"`
	Nodes      []Node      `json:"nodes" yaml:"nodes"`
	UpdatedAt  string      `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
	Attributes interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

type ProfileImportRequest struct {
	Name       string      `json:"name"`
	Source     string      `json:"source"`
	Content    string      `json:"content,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type ProfileImportResponse struct {
	Profile Profile `json:"profile"`
}

type ProfileRefreshRequest struct {
	ProfileID string `json:"profileId"`
}

type ProfileRefreshResponse struct {
	Profile Profile `json:"profile"`
}

type ProfileDeleteRequest struct {
	ProfileID string `json:"profileId"`
}

type ProfileDeleteResponse struct {
	ProfileID string `json:"profileId"`
}

type ProfileListResponse struct {
	Profiles []Profile `json:"profiles"`
}

type ProfileGetResponse struct {
	Profile Profile `json:"profile"`
}
