package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 定义了应用程序的配置结构
type Config struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	Security struct {
		AESKey             string `json:"aes_key"`
		DefaultAdURLBase64 string `json:"default_ad_url_base64"`
	} `json:"security"`
	URLs struct {
		DefaultLivePrefix string `json:"default_live_prefix"`
		HuyaAPIBase       string `json:"huya_api_base"`
		DouyuAPIBase      string `json:"douyu_api_base"`
		YYAPIBase         string `json:"yy_api_base"`
		IptvJsListURL     string `json:"iptv_js_list_url"`
	} `json:"urls"`
	Defaults struct {
		HuyaGID          string `json:"huya_gid"`
		DouyuGID         string `json:"douyu_gid"`
		StreamType       string `json:"stream_type"`
		HuyaCDN          string `json:"huya_cdn"`
		HuyaMedia        string `json:"huya_media"`
		HuyaResponseType string `json:"huya_response_type"`
		BilibiliPlatform string `json:"bilibili_platform"`
		BilibiliQuality  string `json:"bilibili_quality"`
		BilibiliLine     string `json:"bilibili_line"`
		YoutubeQuality   string `json:"youtube_quality"`
		YYQuality        string `json:"yy_quality"`
	} `json:"defaults"`
	TestVideo struct {
		LogoURL      string `json:"logo_url"`
		TimeVideoURL string `json:"time_video_url"`
		TestAdURL1   string `json:"test_ad_url_1"`
		TestAdURL2   string `json:"test_ad_url_2"`
	} `json:"test_video"`
}

// GlobalConfig 用于存储全局配置
var GlobalConfig Config

// LoadConfig 从指定路径加载配置文件
func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法打开配置文件 %s: %v", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&GlobalConfig)
	if err != nil {
		return fmt.Errorf("无法解析配置文件 %s: %v", path, err)
	}

	return nil
}
