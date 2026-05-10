package media

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"sakiko.local/sakiko-core/interfaces"
)

const (
	huluJapanURL          = "https://id.hulu.jp"
	bilibiliTWURLPattern  = "https://api.bilibili.com/pgc/player/web/playurl?avid=50762638&cid=100279344&qn=0&type=&otype=json&ep_id=268176&fourk=1&fnver=0&fnval=16&session=%s&module=bangumi"
	abemaCheckURL         = "https://api.abema.io/v1/ip/check?device=android"
	hboMaxTokenURL        = "https://default.any-any.prd.api.hbomax.com/token?realm=bolt&deviceId=afbb5daa-c327-461d-9460-d8e4b3ee4a1f"
	hboMaxBootstrapURL    = "https://default.any-any.prd.api.hbomax.com/session-context/headwaiter/v1/bootstrap"
	hboMaxPlaybackInfoURL = "https://default.any-any.prd.api.hbomax.com/any/playback/v1/playbackInfo"
	hboMaxPlaybackCookie  = "st=eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJ0b2tlbi0wOWQxOTg4Yy1mZmUzLTQxMDEtOWI5My0yNDU1ZTkyNGQ1YjYiLCJpc3MiOiJmcGEtaXNzdWVyIiwic3ViIjoiVVNFUklEOmJvbHQ6YjYzOTgxZWQtNzA2MC00ZGYwLThkZGItZjA2YjFkNWRjZWVkIiwiaWF0IjoxNzQzODQwMzgwLCJleHAiOjIwNTkyMDAzODAsInR5cGUiOiJBQ0NFU1NfVE9LRU4iLCJzdWJkaXZpc2lvbiI6ImJlYW1fYW1lciIsInNjb3BlIjoiZGVmYXVsdCIsImlpZCI6IjQwYTgzZjNlLTY4OTktNDE3Mi1hMWY2LWJjZDVjN2ZkNjA4NSIsInZlcnNpb24iOiJ2MyIsImFub255bW91cyI6ZmFsc2UsImRldmljZUlkIjoiNWY3YzViZjQtYjc4Ny00NDRjLWJhYTYtMzU5MzgwYWFiM2RmIn0.f5HTgIV2v0nQQDp5LQG0xqLrxyACdvnMDiWO_viX_CUGqtc5ncSjp_LgM30QFkkMnINFhzKEGRpsZvb-o3Pj_Z39uRBr5LCeiCPR7ssV-_SXyRFVRRDEB2lpxyz7jmdD1SxvA06HnEwTbZQzlbZ7g9GXq02yNdEfHlqYEh_4WF88UbXfeieYTd4TH7kwN1RE50NfQUS6f0WmzpAbpiULyd87mpTeynchFNMMz-YHVzZ_-nDW6geihXc3tS0FKVSR8fdOSPQFzEYOLCfhInufiPahiXI-OKF89aShAqM-y4Hx_eukGnsq3mO5wa3unnqVr9Kzc61BIhHh1Hs2bqYiYg;"
	youtubePremiumURL     = "https://www.youtube.com/premium"
	youtubePremiumCookie  = "YSC=BiCUU3-5Gdk; CONSENT=YES+cb.20220301-11-p0.en+FX+700; GPS=1; VISITOR_INFO1_LIVE=4VwPMkB7W5A; PREF=tz=Asia.Shanghai; _gcl_au=1.1.1809531354.1646633279"
	primeVideoURL         = "https://www.primevideo.com"
	primeVideoAPIURL      = "https://ab9f7h23rcdn.eu.api.amazonvideo.com/cdp/appleedge/getDataByTransform/v1/apple/detail/vod/v1.kt?itemId=amzn1.dv.gti.e6b39984-2bb6-f7d0-33e4-08ec574947f0&deviceId=6F97F9CCFA2243F1A3C44BD3C7F7908E&deviceTypeId=A3JTVZS31ZJ340&density=2x&firmware=10.6800.16104.3&format=json&enabledFeatures=denarius.location.gen4.daric.siglos.siglosPartnerBilling.contentDescriptors.contentDescriptorsV2.productPlacement.zeno.seriesSearch.tapsV2.dateTimeLocalization.multiSourcedEvents.mseEventLevelOffers.liveWatchModal.lbv.daapi.maturityRatingDecoration.seasonTrailer.cleanSlate.xbdModalV2.xbdModalVdp.playbackPinV2.exploreTab.reactions.progBadging.atfEpTimeVis.prereleaseCx.vppaConsent.episodicRelease.movieVam.movieVamCatalog&journeyIngressContext=8%7CEgRzdm9k&osLocale=zh_Hans_CN&timeZoneId=Asia%2FShanghai&uxLocale=zh_CN"
	primeVideoAppUA       = "PrimeVideo/10.68 (iPad; iOS 18.3.2; Scale/2.00)"
	tikTokHomeURL         = "https://www.tiktok.com/"
	tikTokRegionURL       = "https://www.tiktok.com/passport/web/store_region/"
	spotifySignupURL      = "https://www.spotify.com/tw/signup"
	steamAppURL           = "https://store.steampowered.com/app/761830"
	chatGPTURL            = "https://chatgpt.com"
	chatGPTIOSURL         = "https://ios.chat.openai.com"
	chatGPTTraceURL       = "https://chatgpt.com/cdn-cgi/trace"
	claudeURL             = "https://claude.ai/"
	geminiExecuteURL      = "https://gemini.google.com/_/BardChatUi/data/batchexecute"
	geminiExecuteBody     = `f.req=[[["K4WWud","[[0],[\"en-US\"]]",null,"generic"]]]`
	mediaAndroidUA        = "Dalvik/2.1.0 (Linux; U; Android 9; ALP-AL00 Build/HUAWEIALP-AL00)"
)

var (
	currentTerritoryRegexp = regexp.MustCompile(`"currentTerritory"\s*:\s*"([^"]+)"`)
	spotifyCountryRegexp   = regexp.MustCompile(`geoCountry":"([^"]+)"`)
	steamCurrencyRegexp    = regexp.MustCompile(`priceCurrency"\s*:\s*"([^"]+)"`)
	youtubeCountryRegexp   = regexp.MustCompile(`"countryCode":"([^"]+)"`)
	geminiLocationRegexp   = regexp.MustCompile(`\[\[\\"([^"]+)\\"\,\\"S`)
)

func probeHuluJP(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformHuluJP, "Hulu Japan")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: huluJapanURL,
		Headers: map[string]string{
			"Accept":          "*/*;q=0.8",
			"Accept-Language": "en-US,en;q=0.5",
			"User-Agent":      mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	lowerFinalURL := strings.ToLower(snapshot.FinalURL)
	switch {
	case snapshot.StatusCode == 403 || strings.Contains(lowerFinalURL, "restrict"):
		result.Status = interfaces.MediaUnlockStatusNo
		result.Region = "JP"
	default:
		result.Status = interfaces.MediaUnlockStatusYes
		result.Region = "JP"
	}
	return finalizeResult(result)
}

func probeBilibiliTW(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformBilibiliTW, "Bilibili Taiwan")

	sessionID, err := randomHex(16)
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	requestURL := fmt.Sprintf(bilibiliTWURLPattern, sessionID)
	snapshot, err := performRequest(ctx, proxy, buildBilibiliRequest(requestURL, "", false))
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	status, parseErr := evaluateBilibiliSnapshot(snapshot)
	if parseErr != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = parseErr.Error()
		return finalizeResult(result)
	}

	result.Status = status
	if status == interfaces.MediaUnlockStatusYes {
		result.Region = "TW"
	}
	return finalizeResult(result)
}

func probeAbema(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformAbema, "Abema")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: abemaCheckURL,
		Headers: map[string]string{
			"User-Agent": mediaAndroidUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	var payload struct {
		Region string `json:"isoCountryCode"`
	}
	if err := json.Unmarshal(snapshot.Body, &payload); err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	result.Region = strings.ToUpper(strings.TrimSpace(payload.Region))
	switch result.Region {
	case "JP":
		result.Status = interfaces.MediaUnlockStatusYes
	case "":
		result.Status = interfaces.MediaUnlockStatusNo
	default:
		result.Status = interfaces.MediaUnlockStatusOverseaOnly
	}
	return finalizeResult(result)
}

func probeHBOMax(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformHBOMax, "HBO Max")

	tokenSnapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: hboMaxTokenURL,
		Headers: map[string]string{
			"User-Agent":      mediaBrowserUA,
			"x-device-info":   "beam/5.0.0 (desktop/desktop; Windows/10; afbb5daa-c327-461d-9460-d8e4b3ee4a1f/da0cdd94-5a39-42ef-aa68-54cbc1b852c3)",
			"x-disco-client":  "WEB:10:beam:5.2.1",
			"Accept-Language": "en-US,en;q=0.9",
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	var tokenPayload struct {
		Data struct {
			Attributes struct {
				Token string `json:"token"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(tokenSnapshot.Body, &tokenPayload); err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}
	token := strings.TrimSpace(tokenPayload.Data.Attributes.Token)
	if token == "" {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = "token missing"
		return finalizeResult(result)
	}

	bootstrapSnapshot, err := performRequest(ctx, proxy, requestSpec{
		Method: httpMethodPost,
		URL:    hboMaxBootstrapURL,
		Headers: map[string]string{
			"Cookie":         "st=" + token,
			"User-Agent":     mediaBrowserUA,
			"x-disco-client": "WEB:10:beam:5.2.1",
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	var bootstrapPayload struct {
		Routing struct {
			Domain     string `json:"domain"`
			Tenant     string `json:"tenant"`
			Env        string `json:"env"`
			HomeMarket string `json:"homeMarket"`
		} `json:"routing"`
	}
	if err := json.Unmarshal(bootstrapSnapshot.Body, &bootstrapPayload); err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}
	if bootstrapPayload.Routing.Domain == "" || bootstrapPayload.Routing.Tenant == "" || bootstrapPayload.Routing.Env == "" || bootstrapPayload.Routing.HomeMarket == "" {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = "routing missing"
		return finalizeResult(result)
	}

	usersURL := fmt.Sprintf(
		"https://default.%s-%s.%s.%s/users/me",
		bootstrapPayload.Routing.Tenant,
		bootstrapPayload.Routing.HomeMarket,
		bootstrapPayload.Routing.Env,
		bootstrapPayload.Routing.Domain,
	)
	userSnapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: usersURL,
		Headers: map[string]string{
			"Cookie":     "st=" + token,
			"User-Agent": mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	var userPayload struct {
		Data struct {
			Attributes struct {
				CurrentLocationTerritory string `json:"currentLocationTerritory"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(userSnapshot.Body, &userPayload); err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	result.Region = strings.ToUpper(strings.TrimSpace(userPayload.Data.Attributes.CurrentLocationTerritory))
	if result.Region == "" {
		result.Status = interfaces.MediaUnlockStatusNo
		return finalizeResult(result)
	}

	availableRegionSnapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: hboMaxPublicURL(),
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if err == nil {
		if regions := extractHBOMaxRegions(availableRegionSnapshot); len(regions) > 0 && !regions[result.Region] {
			result.Status = interfaces.MediaUnlockStatusNo
			return finalizeResult(result)
		}
	}

	playbackSnapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: hboMaxPlaybackInfoURL,
		Headers: map[string]string{
			"Cookie":     hboMaxPlaybackCookie,
			"User-Agent": mediaBrowserUA,
		},
	})
	if err == nil && containsVPNSignal(playbackSnapshot.Body) {
		result.Status = interfaces.MediaUnlockStatusNo
		result.Error = "vpn detected"
		return finalizeResult(result)
	}
	result.Status = interfaces.MediaUnlockStatusYes
	return finalizeResult(result)
}

func probeYouTubePremium(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformYouTubePremium, "YouTube Premium")

	first, err1 := performRequest(ctx, proxy, requestSpec{
		URL: youtubePremiumURL,
		Headers: map[string]string{
			"Accept-Language": "en",
			"Cookie":          youtubePremiumCookie,
			"User-Agent":      mediaBrowserUA,
		},
	})
	second, err2 := performRequest(ctx, proxy, requestSpec{
		URL: youtubePremiumURL,
		Headers: map[string]string{
			"Accept-Language": "en",
			"User-Agent":      mediaBrowserUA,
		},
	})
	if err1 != nil && err2 != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = joinErrors(err1, err2)
		return finalizeResult(result)
	}

	combined := strings.Join([]string{string(first.Body), string(second.Body)}, "\n")
	if strings.Contains(combined, "www.google.cn") {
		result.Status = interfaces.MediaUnlockStatusNo
		result.Region = "CN"
		return finalizeResult(result)
	}

	if matches := youtubeCountryRegexp.FindStringSubmatch(combined); len(matches) > 1 {
		result.Region = strings.ToUpper(strings.TrimSpace(matches[1]))
	}
	if strings.Contains(combined, "purchaseButtonOverride") || strings.Contains(combined, "Start trial") {
		result.Status = interfaces.MediaUnlockStatusYes
		return finalizeResult(result)
	}
	if result.Region != "" {
		result.Status = interfaces.MediaUnlockStatusYes
		return finalizeResult(result)
	}
	result.Status = interfaces.MediaUnlockStatusNo
	return finalizeResult(result)
}

func probePrimeVideo(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformPrimeVideo, "Prime Video")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: primeVideoURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	matches := currentTerritoryRegexp.FindStringSubmatch(string(snapshot.Body))
	if len(matches) < 2 {
		result.Status = interfaces.MediaUnlockStatusUnsupported
		return finalizeResult(result)
	}

	result.Region = strings.ToUpper(strings.TrimSpace(matches[1]))
	apiSnapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: primeVideoAPIURL,
		Headers: map[string]string{
			"User-Agent": primeVideoAppUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}
	if containsVPNSignal(apiSnapshot.Body) {
		result.Status = interfaces.MediaUnlockStatusNo
		result.Error = "vpn detected"
		return finalizeResult(result)
	}

	result.Status = interfaces.MediaUnlockStatusYes
	return finalizeResult(result)
}

func probeTikTok(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformTikTok, "TikTok")

	homeSnapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: tikTokHomeURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	regionSnapshot, regionErr := performRequest(ctx, proxy, requestSpec{
		Method: httpMethodPost,
		URL:    tikTokRegionURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if regionErr == nil {
		var payload struct {
			Data struct {
				StoreRegion string `json:"store_region"`
			} `json:"data"`
		}
		if err := json.Unmarshal(regionSnapshot.Body, &payload); err == nil {
			result.Region = strings.ToUpper(strings.TrimSpace(payload.Data.StoreRegion))
		}
	}

	lowerFinalURL := strings.ToLower(homeSnapshot.FinalURL)
	if strings.Contains(lowerFinalURL, "/about") || strings.Contains(lowerFinalURL, "/status") || strings.Contains(lowerFinalURL, "landing") {
		result.Status = interfaces.MediaUnlockStatusNo
		if strings.EqualFold(result.Region, "CN") {
			result.Display = "Douyin (CN)"
		}
		return finalizeResult(result)
	}

	result.Status = interfaces.MediaUnlockStatusYes
	return finalizeResult(result)
}

func probeSpotify(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformSpotify, "Spotify")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: spotifySignupURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	matches := spotifyCountryRegexp.FindStringSubmatch(string(snapshot.Body))
	if len(matches) < 2 {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = "country missing"
		return finalizeResult(result)
	}
	result.Status = interfaces.MediaUnlockStatusYes
	result.Region = strings.ToUpper(strings.TrimSpace(matches[1]))
	result.Display = "Region (" + result.Region + ")"
	return finalizeResult(result)
}

func probeSteam(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformSteam, "Steam")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: steamAppURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	matches := steamCurrencyRegexp.FindStringSubmatch(string(snapshot.Body))
	if len(matches) < 2 {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = "currency missing"
		return finalizeResult(result)
	}
	result.Status = interfaces.MediaUnlockStatusYes
	result.Region = strings.ToUpper(strings.TrimSpace(matches[1]))
	result.Display = "Currency (" + result.Region + ")"
	return finalizeResult(result)
}

func probeChatGPT(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformChatGPT, "ChatGPT")

	headSnapshot, headErr := performChatGPTRedirectProbe(ctx, proxy)
	traceSnapshot, traceErr := performRequest(ctx, proxy, requestSpec{
		URL: chatGPTTraceURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if traceErr == nil {
		result.Region = parseTraceValue(string(traceSnapshot.Body), "loc")
	}

	iosSnapshot, iosErr := performRequest(ctx, proxy, requestSpec{
		URL: chatGPTIOSURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if iosErr != nil && traceErr != nil && headErr != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = joinErrors(iosErr, traceErr, headErr)
		return finalizeResult(result)
	}
	if headErr != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = headErr.Error()
		return finalizeResult(result)
	}
	if iosErr != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = joinErrors(iosErr, traceErr)
		return finalizeResult(result)
	}

	result.Status, result.Error = evaluateChatGPTProbe(result.Region, chatGPTRedirectDetected(headSnapshot), iosSnapshot.Body)
	return finalizeResult(result)
}

func probeClaude(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformClaude, "Claude")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		URL: claudeURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	status, errText := evaluateClaudeSnapshot(snapshot)
	result.Status = status
	result.Error = errText
	if result.Status == interfaces.MediaUnlockStatusFailed {
		return finalizeResult(result)
	}
	return finalizeResult(result)
}

func probeGemini(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := newProbeResult(interfaces.MediaUnlockPlatformGemini, "Gemini")

	snapshot, err := performRequest(ctx, proxy, requestSpec{
		Method: httpMethodPost,
		URL:    geminiExecuteURL,
		Headers: map[string]string{
			"Accept-Language": "en-US",
			"Content-Type":    "application/x-www-form-urlencoded;charset=UTF-8",
			"User-Agent":      mediaBrowserUA,
		},
		Body: []byte(geminiExecuteBody),
	})
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	matches := geminiLocationRegexp.FindStringSubmatch(string(snapshot.Body))
	if len(matches) < 2 {
		result.Status = interfaces.MediaUnlockStatusUnsupported
		return finalizeResult(result)
	}

	result.Status = interfaces.MediaUnlockStatusYes
	result.Region = strings.TrimSpace(matches[1])
	return finalizeResult(result)
}

func newProbeResult(platform interfaces.MediaUnlockPlatform, name string) interfaces.MediaUnlockPlatformResult {
	return interfaces.MediaUnlockPlatformResult{
		Platform: platform,
		Name:     name,
		Mode:     interfaces.MediaUnlockModeUnknown,
	}
}

func finalizeResult(result interfaces.MediaUnlockPlatformResult) interfaces.MediaUnlockPlatformResult {
	if strings.TrimSpace(result.Display) == "" {
		result.Display = defaultMediaDisplay(result)
	}
	return result
}

func defaultMediaDisplay(result interfaces.MediaUnlockPlatformResult) string {
	region := strings.TrimSpace(result.Region)
	errText := strings.TrimSpace(result.Error)

	switch result.Status {
	case interfaces.MediaUnlockStatusYes:
		if region != "" {
			return "Yes (Region: " + region + ")"
		}
		return "Yes"
	case interfaces.MediaUnlockStatusNo:
		return buildNoDisplay(region, errText)
	case interfaces.MediaUnlockStatusOriginalsOnly:
		if region != "" {
			return "Originals Only (Region: " + region + ")"
		}
		return "Originals Only"
	case interfaces.MediaUnlockStatusWebOnly:
		if errText != "" && region != "" {
			return "Web Only (" + humanizeMediaReason(errText) + ";Region: " + region + ")"
		}
		if errText != "" {
			return "Web Only (" + humanizeMediaReason(errText) + ")"
		}
		if region != "" {
			return "Web Only (Region: " + region + ")"
		}
		return "Web Only"
	case interfaces.MediaUnlockStatusOverseaOnly:
		if region != "" {
			return "Oversea Only (Region: " + region + ")"
		}
		return "Oversea Only"
	case interfaces.MediaUnlockStatusUnsupported:
		if region != "" {
			return "Unsupported (Region: " + region + ")"
		}
		return "Unsupported"
	case interfaces.MediaUnlockStatusFailed:
		if errText != "" {
			return "Failed (" + errText + ")"
		}
		return "Failed"
	default:
		if errText != "" {
			return errText
		}
		return "-"
	}
}

func buildNoDisplay(region string, errText string) string {
	switch strings.ToLower(strings.TrimSpace(errText)) {
	case "vpn detected":
		if region != "" {
			return "No (VPN Detected;Region: " + region + ")"
		}
		return "No (VPN Detected)"
	case "blocked":
		return "No (Blocked)"
	case "unsupported region":
		return "No (Unsupported Region)"
	}

	if errText != "" && region != "" {
		return "No (" + humanizeMediaReason(errText) + ";Region: " + region + ")"
	}
	if errText != "" {
		return "No (" + humanizeMediaReason(errText) + ")"
	}
	if region != "" {
		return "No (Region: " + region + ")"
	}
	return "No"
}

func extractHBOMaxRegions(snapshot httpSnapshot) map[string]bool {
	regions := make(map[string]bool)
	for _, match := range regexp.MustCompile(`"/[a-z]{2}/[a-z]{2}"`).FindAllString(string(snapshot.Body), -1) {
		parts := strings.Split(strings.Trim(match, `"`), "/")
		if len(parts) >= 3 {
			regions[strings.ToUpper(strings.TrimSpace(parts[2]))] = true
		}
	}
	return regions
}

func containsVPNSignal(body []byte) bool {
	text := strings.ToLower(string(body))
	return strings.Contains(text, "vpn") || strings.Contains(text, "proxy")
}

func hboMaxPublicURL() string {
	return "https://www.hbomax.com/"
}

func humanizeMediaReason(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "disallowed isp[1]":
		return "Disallowed ISP[1]"
	case "disallowed isp[2]":
		return "Disallowed ISP[2]"
	case "vpn detected":
		return "VPN Detected"
	case "unsupported region":
		return "Unsupported Region"
	case "page error":
		return "Page Error"
	default:
		return value
	}
}

func performChatGPTRedirectProbe(ctx context.Context, proxy interfaces.Vendor) (httpSnapshot, error) {
	spec := requestSpec{
		Method:  http.MethodHead,
		URL:     chatGPTURL,
		NoRedir: true,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	}

	snapshot, err := performRequest(ctx, proxy, spec)
	if err == nil {
		return snapshot, nil
	}

	spec.Method = http.MethodGet
	return performRequest(ctx, proxy, spec)
}

func chatGPTRedirectDetected(snapshot httpSnapshot) bool {
	if len(snapshot.Headers.Values("Location")) > 0 {
		return true
	}
	location := strings.TrimSpace(snapshot.Headers.Get("Location"))
	return snapshot.StatusCode >= 300 && snapshot.StatusCode < 400 && location != ""
}

func evaluateChatGPTProbe(region string, redirectReady bool, body []byte) (interfaces.MediaUnlockStatus, string) {
	text := strings.ToLower(string(body))
	cfDetails := strings.ToLower(extractChatGPTCFDetails(body))
	region = strings.ToUpper(strings.TrimSpace(region))

	switch {
	case strings.Contains(text, "blocked_why_headline"):
		return interfaces.MediaUnlockStatusNo, "blocked"
	case strings.Contains(text, "unsupported_country_region_territory"):
		return interfaces.MediaUnlockStatusNo, "unsupported region"
	case strings.Contains(cfDetails, "(1)") || strings.Contains(text, "(1)"):
		if redirectReady {
			return interfaces.MediaUnlockStatusWebOnly, "disallowed isp[1]"
		}
		return interfaces.MediaUnlockStatusNo, "disallowed isp[1]"
	case strings.Contains(cfDetails, "(2)") || strings.Contains(text, "(2)"):
		if redirectReady {
			return interfaces.MediaUnlockStatusWebOnly, "disallowed isp[2]"
		}
		return interfaces.MediaUnlockStatusNo, "disallowed isp[2]"
	default:
		if region != "" || redirectReady {
			return interfaces.MediaUnlockStatusYes, ""
		}
		return interfaces.MediaUnlockStatusNo, ""
	}
}

func extractChatGPTCFDetails(body []byte) string {
	var payload struct {
		CFDetails any `json:"cf_details"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	switch value := payload.CFDetails.(type) {
	case string:
		return value
	case nil:
		return ""
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}

func evaluateClaudeSnapshot(snapshot httpSnapshot) (interfaces.MediaUnlockStatus, string) {
	lowerFinalURL := strings.ToLower(snapshot.FinalURL)
	lowerBody := strings.ToLower(string(snapshot.Body))

	if strings.Contains(lowerFinalURL, "unavailable") || strings.Contains(lowerBody, "unavailable") {
		return interfaces.MediaUnlockStatusNo, ""
	}
	if snapshot.StatusCode >= 500 {
		return interfaces.MediaUnlockStatusFailed, fmt.Sprintf("status code: %d", snapshot.StatusCode)
	}
	return interfaces.MediaUnlockStatusYes, ""
}

func parseTraceValue(raw string, key string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, key+"=") {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(line, key+"="))
	}
	return ""
}

const httpMethodPost = "POST"
