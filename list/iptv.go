package list

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type ChannelList struct {
	Data []ChannelInfo `json:"data"`
}

type ChannelInfo struct {
	ChnName string `json:"chnName"`
	PlayUrl string `json:"playUrl"`
}

type PlayUrlResponse struct {
	U string `json:"u"`
}

func getGroupInfo(chnName string) string {
	if strings.ContainsAny(chnName, "少儿卡通 CCTV-14") {
		return "Kids"
	}
	if strings.Contains(chnName, "CCTV") || strings.Contains(chnName, "CGTN") {
		return "CCTV"
	}
	if strings.ContainsAny(chnName, "江苏南京") {
		return "Jiangsu"
	}
	if strings.Contains(chnName, "卫视") {
		return "Weishi"
	}
	if strings.ContainsAny(chnName, "CETV 教育") {
		return "Education"
	}
	return "Other"
}

func GetIptvJs() (string, error) {
	listUrl := "http://live.epg.gitv.tv/tagNewestEpgList/JS_CUCC/1/100/0.json"
	resp, err := http.Get(listUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch channel list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch channel list: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var channelList ChannelList
	if err := json.Unmarshal(body, &channelList); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	m3uData := "#EXTM3U\n"

	for _, item := range channelList.Data {
		chnName := item.ChnName
		playUrl := item.PlayUrl

		playResp, err := http.Get(playUrl)
		if err != nil {
			fmt.Printf("Error fetching play URL for %s: %v\n", chnName, err)
			continue
		}
		playBody, err := io.ReadAll(playResp.Body)
		playResp.Body.Close()

		if err != nil || playResp.StatusCode != http.StatusOK {
			fmt.Printf("Error reading play URL for %s: %v\n", chnName, err)
			continue
		}

		var playData PlayUrlResponse
		if err := json.Unmarshal(playBody, &playData); err != nil {
			fmt.Printf("Error unmarshalling play URL for %s: %v\n", chnName, err)
			continue
		}

		playUrlReal := playData.U
		playUrlReal += "?" + playResp.Request.URL.RawQuery

		groupName := getGroupInfo(chnName)
		re := regexp.MustCompile(`高清 | 超清 | 蓝光|-8M|-`)
		tvgName := re.ReplaceAllString(chnName, "")

		m3uData += fmt.Sprintf("#EXTINF:-1 group-title=\"%s\" tvg-name=\"%s\",%s\n", groupName, tvgName, chnName)
		m3uData += fmt.Sprintf("%s\n", playUrlReal)
	}

	return m3uData, nil
}
