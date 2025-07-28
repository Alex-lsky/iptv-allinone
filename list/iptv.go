package list

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// ChannelList represents the top-level structure of the IPTV channel list API response.
type ChannelList struct {
	Data []ChannelInfo `json:"data"`
}

// ChannelInfo represents the information for a single channel.
type ChannelInfo struct {
	Tag       string `json:"tag"`
	ChnunCode string `json:"chnunCode"`
	ChnName   string `json:"chnName"`
	ChnCode   string `json:"chnCode"`
	PlayUrl   string `json:"playUrl"`
}

// PlayUrlResponse represents the structure of the response from a channel's PlayUrl.
type PlayUrlResponse struct {
	U string `json:"u"`
}

// getGroupInfo determines the group title for a channel based on its name.
// This replicates the logic from the Python script's get_group_info function.
func getGroupInfo(chnName string) string {
	groups := map[string][]string{
		"少儿":   {"少儿", "卡通", "CCTV-14"},
		"CCTV": {"CCTV", "CGTN"},
		"江苏":   {"江苏", "南京"},
		"卫视":   {"卫视"},
		"教育":   {"CETV", "教育"},
	}

	for key, keywords := range groups {
		for _, keyword := range keywords {
			if strings.Contains(chnName, keyword) {
				return key
			}
		}
	}

	return "其他"
}

// GetIptvJs fetches the IPTV channel list, resolves real play URLs,
// and generates an M3U playlist string.
// This function replicates the core logic of the provided Python script.
func GetIptvJs() (string, error) {
	// 1. Fetch the main channel list
	listUrl := "http://live.epg.gitv.tv/tagNewestEpgList/JS_CUCC/1/100/0.json"
	resp, err := http.Get(listUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch channel list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch channel list, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read channel list response body: %w", err)
	}

	var channelList ChannelList
	if err := json.Unmarshal(body, &channelList); err != nil {
		return "", fmt.Errorf("failed to unmarshal channel list JSON: %w", err)
	}

	// 2. Start building the M3U content
	m3uDataFull := "#EXTM3U\n"
	m3uDataKid := "#EXTM3U\n"

	// 3. Process each channel
	for _, item := range channelList.Data {
		chnName := item.ChnName
		playUrl := item.PlayUrl

		// 3a. Fetch the real play URL for the channel
		playResp, err := http.Get(playUrl)
		if err != nil {
			// Log error and skip this channel if fetching play URL fails
			fmt.Printf("Error fetching play URL for %s (%s): %v\n", chnName, playUrl, err)
			continue
		}
		playBody, err := io.ReadAll(playResp.Body)
		playResp.Body.Close() // Close immediately after reading

		if err != nil {
			fmt.Printf("Error reading play URL response for %s: %v\n", chnName, err)
			continue
		}

		if playResp.StatusCode != http.StatusOK {
			fmt.Printf("Non-200 status code for play URL of %s: %d\n", chnName, playResp.StatusCode)
			continue
		}

		var playData PlayUrlResponse
		if err := json.Unmarshal(playBody, &playData); err != nil {
			fmt.Printf("Error unmarshalling play URL response for %s: %v\n", chnName, err)
			continue
		}

		playUrlReal := playData.U

		// 3b. Determine group and tvg-name
		groupName := getGroupInfo(chnName)

		// Replicate Python's tvgName logic: remove "高清", "超清", "超清", "-8M", "-"
		re := regexp.MustCompile(`高清|超清|超清|-8M|-`)
		tvgName := re.ReplaceAllString(chnName, "")

		// 3c. Add entry to the full M3U list
		m3uDataFull += fmt.Sprintf("#EXTINF:-1 group-title=\"%s\" tvg-name=\"%s\",%s\n", groupName, tvgName, chnName)
		m3uDataFull += fmt.Sprintf("%s\n", playUrlReal)

		// 3d. Add entry to the kid-friendly M3U list (exclude 少儿 and 其他 groups)
		if groupName != "少儿" && groupName != "其他" {
			m3uDataKid += fmt.Sprintf("#EXTINF:-1 group-title=\"%s\",%s\n", groupName, chnName)
			m3uDataKid += fmt.Sprintf("%s\n", playUrlReal)
		}
	}

	// For now, we will only return the full list.
	// In a more advanced version, we could handle custom M3U files like the Python script.
	// The Python script also saves to files; this Go version is designed to return the string
	// for the web handler to serve directly.

	// Returning the full list as the primary output for the M3U endpoint.
	return m3uDataFull, nil
}
