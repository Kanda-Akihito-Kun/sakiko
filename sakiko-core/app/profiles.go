package app

import (
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func (s *Service) ListProfileSummaries() []ProfileSummary {
	if s == nil || s.api == nil {
		return []ProfileSummary{}
	}
	profiles := s.api.ListProfiles().Profiles
	summaries := make([]ProfileSummary, 0, len(profiles))
	for _, profile := range profiles {
		summaries = append(summaries, BuildProfileSummary(profile))
	}
	return summaries
}

func BuildProfileSummary(profile interfaces.Profile) ProfileSummary {
	return ProfileSummary{
		ID:             profile.ID,
		Name:           profile.Name,
		Source:         profile.Source,
		UpdatedAt:      profile.UpdatedAt,
		NodeCount:      len(profile.Nodes),
		RemainingBytes: ProfileRemainingBytes(profile.Attributes),
		ExpiresAt:      ProfileExpiresAt(profile.Attributes),
	}
}

func ProfileRemainingBytes(attributes any) uint64 {
	info := subscriptionUserinfo(attributes)
	if remaining, ok := uint64AttrValue(info["remaining"]); ok {
		return remaining
	}
	total, totalOK := uint64AttrValue(info["total"])
	if !totalOK {
		return 0
	}
	upload, _ := uint64AttrValue(info["upload"])
	download, _ := uint64AttrValue(info["download"])
	if total <= upload+download {
		return 0
	}
	return total - upload - download
}

func ProfileExpiresAt(attributes any) string {
	info := subscriptionUserinfo(attributes)
	if expiresAt, ok := stringAttrValue(info["expiresAt"]); ok && strings.TrimSpace(expiresAt) != "" {
		return expiresAt
	}
	if expire, ok := uint64AttrValue(info["expire"]); ok && expire > 0 {
		return time.Unix(int64(expire), 0).UTC().Format(time.RFC3339)
	}
	return ""
}

func subscriptionUserinfo(attributes any) map[string]any {
	root, ok := attributes.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	value, ok := root["subscriptionUserinfo"]
	if !ok {
		return map[string]any{}
	}
	info, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return info
}

func uint64AttrValue(value any) (uint64, bool) {
	switch typed := value.(type) {
	case uint64:
		return typed, true
	case uint:
		return uint64(typed), true
	case int:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case int64:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case float64:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	default:
		return 0, false
	}
}

func stringAttrValue(value any) (string, bool) {
	typed, ok := value.(string)
	return typed, ok
}
