package interfaces

type DownloadTargetSource string

const (
	DownloadTargetSourceCloudflare DownloadTargetSource = "cloudflare"
	DownloadTargetSourceSpeedtest  DownloadTargetSource = "speedtest"
)

type DownloadTarget struct {
	ID          string               `json:"id"`
	Source      DownloadTargetSource `json:"source"`
	Name        string               `json:"name"`
	City        string               `json:"city,omitempty"`
	Country     string               `json:"country,omitempty"`
	CountryCode string               `json:"countryCode,omitempty"`
	Sponsor     string               `json:"sponsor,omitempty"`
	Host        string               `json:"host,omitempty"`
	Endpoint    string               `json:"endpoint,omitempty"`
	DownloadURL string               `json:"downloadURL"`
}

type DownloadTargetListResponse struct {
	Targets []DownloadTarget `json:"targets"`
}

func DefaultDownloadTarget() DownloadTarget {
	return DownloadTarget{
		ID:          "cloudflare-default",
		Source:      DownloadTargetSourceCloudflare,
		Name:        "Cloudflare Default",
		City:        "Global",
		Country:     "Default",
		Sponsor:     "Cloudflare",
		Host:        "speed.cloudflare.com",
		DownloadURL: defaultDownloadURL,
	}
}
