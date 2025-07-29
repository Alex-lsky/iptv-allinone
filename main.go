// Package Golang
// @Time:2024/02/20 21:30
// @File:main.go
// @SoftWare:Goland

package main

import (
	"Golang/list"
	"Golang/liveurls"
	"Golang/proxy" // Import the new proxy package
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/forgoer/openssl"
	"github.com/gin-gonic/gin"
)

func duanyan(adurl string, realurl any) string {
	var liveurl string
	if str, ok := realurl.(string); ok {
		liveurl = str
	} else {
		liveurl = adurl
	}
	return liveurl
}

func getTestVideoUrl(c *gin.Context) {
	TimeLocation, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		TimeLocation = time.FixedZone("CST", 8*60*60)
	}
	str_time := time.Now().In(TimeLocation).Format("2006-01-02 15:04:05")
	fmt.Fprintln(c.Writer, "#EXTM3U")
	fmt.Fprintln(c.Writer, "#EXTINF:-1 tvg-name=\""+str_time+"\" tvg-logo=\""+GlobalConfig.TestVideo.LogoURL+"\" group-title=\"列表更新时间\","+str_time)
	fmt.Fprintln(c.Writer, GlobalConfig.TestVideo.TimeVideoURL)
	fmt.Fprintln(c.Writer, "#EXTINF:-1 tvg-name=\"4K60PSDR-H264-AAC测试\" tvg-logo=\""+GlobalConfig.TestVideo.LogoURL+"\" group-title=\"4K频道\",4K60PSDR-H264-AAC测试")
	fmt.Fprintln(c.Writer, GlobalConfig.TestVideo.TestAdURL1)
	fmt.Fprintln(c.Writer, "#EXTINF:-1 tvg-name=\"4K60PHLG-HEVC-EAC3测试\" tvg-logo=\""+GlobalConfig.TestVideo.LogoURL+"\" group-title=\"4K频道\",4K60PHLG-HEVC-EAC3测试")
	fmt.Fprintln(c.Writer, GlobalConfig.TestVideo.TestAdURL2)
}

func getLivePrefix(c *gin.Context) string {
	firstUrl := c.DefaultQuery("url", GlobalConfig.URLs.DefaultLivePrefix)
	realUrl, _ := url.QueryUnescape(firstUrl)
	return realUrl
}

func setupRouter(adurl string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.HEAD("/", func(c *gin.Context) {
		c.String(http.StatusOK, "请求成功！")
	})

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "请求成功！")
	})

	r.GET("/huyayqk.m3u", func(c *gin.Context) {
		yaobj := &list.HuyaYqk{}
		res, _ := yaobj.HuYaYqk(GlobalConfig.URLs.HuyaAPIBase + "?iGid=" + GlobalConfig.Defaults.HuyaGID)
		var result list.YaResponse
		json.Unmarshal(res, &result)
		pageCount := result.ITotalPage
		pageSize := result.IPageSize
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=huyayqk.m3u")
		getTestVideoUrl(c)

		for i := 1; i <= pageCount; i++ {
			apiRes, _ := yaobj.HuYaYqk(fmt.Sprintf("%s?iGid=%s&iPageNo=%d&iPageSize=%d", GlobalConfig.URLs.HuyaAPIBase, GlobalConfig.Defaults.HuyaGID, i, pageSize))
			var res list.YaResponse
			json.Unmarshal(apiRes, &res)
			data := res.VList
			for _, value := range data {
				fmt.Fprintf(c.Writer, "#EXTINF:-1 tvg-logo=\"%s\" group-title=\"%s\", %s\n", value.SAvatar180, value.SGameFullName, value.SNick)
				fmt.Fprintf(c.Writer, "%s/huya/%v\n", getLivePrefix(c), value.LProfileRoom)
			}
		}
	})

	r.GET("/douyuyqk.m3u", func(c *gin.Context) {
		yuobj := &list.DouYuYqk{}
		resAPI, _ := yuobj.Douyuyqk(GlobalConfig.URLs.DouyuAPIBase + "/list")

		var result list.DouYuResponse
		json.Unmarshal(resAPI, &result)
		pageCount := result.Data.Pgcnt

		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=douyuyqk.m3u")
		getTestVideoUrl(c)

		for i := 1; i <= pageCount; i++ {
			apiRes, _ := yuobj.Douyuyqk(GlobalConfig.URLs.DouyuAPIBase + "/" + strconv.Itoa(i))

			var res list.DouYuResponse
			json.Unmarshal(apiRes, &res)
			data := res.Data.Rl

			for _, value := range data {
				fmt.Fprintf(c.Writer, "#EXTINF:-1 tvg-logo=\"https://apic.douyucdn.cn/upload/%s_big.jpg\" group-title=\"%s\", %s\n", value.Av, value.C2name, value.Nn)
				fmt.Fprintf(c.Writer, "%s/douyu/%v\n", getLivePrefix(c), value.Rid)
			}
		}
	})

	r.GET("/yylunbo.m3u", func(c *gin.Context) {
		yylistobj := &list.Yylist{}
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=yylunbo.m3u")
		getTestVideoUrl(c)

		i := 1
		for {
			// Note: The YY API URL is very long and complex. For now, we'll keep it as is,
			// but in a future enhancement, we could break it down further.
			// For now, we'll only replace the base URL and the page parameter.
			// Replace the base URL in the original string
			originalURL := GlobalConfig.URLs.YYAPIBase + "?channel=appstore&compAppid=yymip&exposured=80&hdid=8dce117c5c963bf9e7063e7cc4382178498f8765&hostVersion=8.25.0&individualSwitch=1&ispType=2&netType=2&openCardLive=1&osVersion=16.5&page=%d&stype=2&supportSwan=0&uid=1834958700&unionVersion=0&y0=8b799811753625ef70dbc1cc001e3a1f861c7f0261d4f7712efa5ea232f4bd3ce0ab999309cac0d7869449a56b44c774&y1=8b799811753625ef70dbc1cc001e3a1f861c7f0261d4f7712efa5ea232f4bd3ce0ab999309cac0d7869449a56b44c774&y11=9c03c7008d1fdae4873436607388718b&y12=9d8393ec004d98b7e20f0c347c3a8c24&yv=1&yyVersion=8.25.0"
			apiRes := yylistobj.Yylb(fmt.Sprintf(originalURL, i))
			var res list.ApiResponse
			json.Unmarshal([]byte(apiRes), &res)
			for _, value := range res.Data.Data {
				fmt.Fprintf(c.Writer, "#EXTINF:-1 tvg-logo=\"%s\" group-title=\"%s\", %s\n", value.Avatar, value.Biz, value.Desc)
				fmt.Fprintf(c.Writer, "%s/yy/%v\n", getLivePrefix(c), value.Sid)
			}
			if res.Data.IsLastPage == 1 {
				break
			}
			i++
		}
	})

	// New route for IPTV JS M3U list
	r.GET("/iptv.m3u", func(c *gin.Context) {
		// Update the list/iptv.go to use the config value for the URL
		// For now, we'll pass the URL as a parameter or set it as a global variable there.
		// This requires modifying list/iptv.go as well, which is outside the scope of this diff.
		// As a temporary solution, we will leave the URL as is in the list/iptv.go file.
		// A more complete solution would involve passing the config to the list package.

		// Call the function in list/iptv.go to generate the M3U content
		m3uContent, err := list.GetIptvJs()
		if err != nil {
			// If there's an error, return a 500 Internal Server Error
			c.String(http.StatusInternalServerError, "Failed to generate IPTV list: %v", err)
			return
		}

		// Set the response headers for M3U file download
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=iptv.m3u")

		// Write the generated M3U content to the response
		c.String(http.StatusOK, m3uContent)
	})

	// New route for Proxied IPTV JS M3U list
	r.GET("/proxy.m3u", func(c *gin.Context) {
		// Call the function in list/iptv.go to generate the original M3U content
		originalM3uContent, err := list.GetIptvJs()
		if err != nil {
			// If there's an error, return a 500 Internal Server Error
			c.String(http.StatusInternalServerError, "Failed to generate original IPTV list: %v", err)
			return
		}

		// Parse the original M3U content and replace URLs
		// We'll implement a simple line-by-line replacement for now.
		// A more robust M3U parser could be used in the future.
		lines := strings.Split(originalM3uContent, "\n")
		var newLines []string

		for _, line := range lines {
			// Check if the line is a URL (not starting with #)
			if !strings.HasPrefix(line, "#") && strings.HasPrefix(line, "http") {
				// This is a stream URL, replace it with our proxy URL
				// We need to URL-encode the original URL to pass it as a query parameter
				encodedOriginalURL := url.QueryEscape(line)
				proxyURL := fmt.Sprintf("%s/proxy/stream?url=%s", GlobalConfig.ProxyAddress, encodedOriginalURL)
				newLines = append(newLines, proxyURL)
			} else {
				// Keep the line as is (e.g., #EXTM3U, #EXTINF, empty lines)
				newLines = append(newLines, line)
			}
		}

		// Join the modified lines back into a single string
		modifiedM3uContent := strings.Join(newLines, "\n")

		// Set the response headers for M3U file download
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=proxy.m3u")

		// Write the modified M3U content to the response
		c.String(http.StatusOK, modifiedM3uContent)
	})

	// New route for proxying individual streams
	r.GET("/proxy/stream", func(c *gin.Context) {
		// Get the original URL from the query parameter
		originalURL := c.Query("url")
		if originalURL == "" {
			c.String(http.StatusBadRequest, "Missing 'url' query parameter")
			return
		}

		// Create an HTTP client for the proxy
		// In a production environment, you might want to configure timeouts, etc.
		client := &http.Client{}

		// Call the ProxyStream function from the proxy package
		// Pass the client, original URL, and the response writer
		err := proxy.ProxyStream(client, originalURL, c.Writer)
		if err != nil {
			// The ProxyStream function already handles sending HTTP errors
			// Log the error for debugging purposes
			log.Printf("Error proxying stream from %s: %v", originalURL, err)
			// We don't need to send another response here as it's already handled
			return
		}
		// If there's no error, the stream has been successfully proxied
		// and the response has been sent to the client.
	})

	r.GET("/:path/:rid", func(c *gin.Context) {
		path := c.Param("path")
		rid := c.Param("rid")
		switch path {
		case "douyin":
			douyinobj := &liveurls.Douyin{}
			douyinobj.Rid = rid
			douyinobj.Stream = c.DefaultQuery("stream", GlobalConfig.Defaults.StreamType)
			c.Redirect(http.StatusMovedPermanently, duanyan(adurl, douyinobj.GetDouYinUrl()))
		case "douyu":
			douyuobj := &liveurls.Douyu{}
			douyuobj.Rid = rid
			douyuobj.Stream_type = c.DefaultQuery("stream", GlobalConfig.Defaults.StreamType)
			c.Redirect(http.StatusMovedPermanently, duanyan(adurl, douyuobj.GetRealUrl()))
		case "huya":
			huyaobj := &liveurls.Huya{}
			huyaobj.Rid = rid
			huyaobj.Cdn = c.DefaultQuery("cdn", GlobalConfig.Defaults.HuyaCDN)
			huyaobj.Media = c.DefaultQuery("media", GlobalConfig.Defaults.HuyaMedia)
			huyaobj.Type = c.DefaultQuery("type", GlobalConfig.Defaults.HuyaResponseType)
			if huyaobj.Type == "display" {
				c.JSON(200, huyaobj.GetLiveUrl())
			} else {
				c.Redirect(http.StatusMovedPermanently, duanyan(adurl, huyaobj.GetLiveUrl()))
			}
		case "bilibili":
			biliobj := &liveurls.BiliBili{}
			biliobj.Rid = rid
			biliobj.Platform = c.DefaultQuery("platform", GlobalConfig.Defaults.BilibiliPlatform)
			biliobj.Quality = c.DefaultQuery("quality", GlobalConfig.Defaults.BilibiliQuality)
			biliobj.Line = c.DefaultQuery("line", GlobalConfig.Defaults.BilibiliLine)
			c.Redirect(http.StatusMovedPermanently, duanyan(adurl, biliobj.GetPlayUrl()))
		case "youtube":
			ytbObj := &liveurls.Youtube{}
			ytbObj.Rid = rid
			ytbObj.Quality = c.DefaultQuery("quality", GlobalConfig.Defaults.YoutubeQuality)
			c.Redirect(http.StatusMovedPermanently, duanyan(adurl, ytbObj.GetLiveUrl()))
		case "yy":
			yyObj := &liveurls.Yy{}
			yyObj.Rid = rid
			yyObj.Quality = c.DefaultQuery("quality", GlobalConfig.Defaults.YYQuality)
			c.Redirect(http.StatusMovedPermanently, duanyan(adurl, yyObj.GetLiveUrl()))
		}
	})
	return r
}

func main() {
	// Load configuration
	if err := LoadConfig("config.json"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	key := []byte(GlobalConfig.Security.AESKey)
	defstr, _ := base64.StdEncoding.DecodeString(GlobalConfig.Security.DefaultAdURLBase64)
	defurl, _ := openssl.AesECBDecrypt(defstr, key, openssl.PKCS7_PADDING)
	r := setupRouter(string(defurl))
	r.Run(GlobalConfig.Server.Port)
}
