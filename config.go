package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 瀹氫箟浜嗗簲鐢ㄧ▼搴忕殑閰嶇疆缁撴瀯
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
	// ProxyEnabled enables or disables the proxy functionality
	ProxyEnabled bool `json:"proxy_enabled"`
	// ProxyAddress is the public address of the proxy server
	ProxyAddress string `json:"proxy_address"`
	// Cache 缂撳瓨閰嶇疆
	Cache CacheConfig `json:"cache"`
}

// CacheConfig 缂撳瓨閰嶇疆缁撴瀯
type CacheConfig struct {
	EnableCache       bool  `json:"enable_cache"`
	M3U8PreloadCount  int   `json:"m3u8_preload_count"`
	SegmentBufferSize int64 `json:"segment_buffer_size"`
	StreamBufferSize  int   `json:"stream_buffer_size"`
	PreloadTimeout    int   `json:"preload_timeout"`
	MaxRetries        int   `json:"max_retries"`
	RetryDelay        int   `json:"retry_delay"`
	// 棰戦亾缂撳瓨閰嶇疆
	ChannelCacheEnabled     bool `json:"channel_cache_enabled"`
	ChannelCacheMaxMemoryMB int  `json:"channel_cache_max_memory_mb"` // 棰戦亾缂撳瓨鏈€澶у唴瀛?(MB)
}

// GlobalConfig 鐢ㄤ簬瀛樺偍鍏ㄥ眬閰嶇疆
var GlobalConfig Config

// LoadConfig 浠庢寚瀹氳矾寰勫姞杞介厤缃枃浠?
func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("鏃犳硶鎵撳紑閰嶇疆鏂囦欢 %s: %v", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&GlobalConfig)
	if err != nil {
		return fmt.Errorf("鏃犳硶瑙ｆ瀽閰嶇疆鏂囦欢 %s: %v", path, err)
	}

	return nil
}
