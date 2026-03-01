package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProxyStream handles the core logic for proxying an HTTP stream with caching support.
// It takes the original upstream URL, fetches the stream with buffering/caching, and pipes it to the client's response writer.
// It also takes the proxyAddress to rewrite M3U8 playlists if necessary.
func ProxyStream(c *http.Client, originalURL string, w http.ResponseWriter, proxyAddress string) error {
	// 使用全局配置中的缓存配置
	config := GlobalCacheConfig
	if config == nil {
		// 如果没有设置全局配置，使用默认配置
		config = DefaultCacheConfig()
	}

	// 检查是否是 M3U8 流
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		log.Printf("Error parsing originalURL %s: %v", originalURL, err)
		http.Error(w, "Failed to parse stream URL", http.StatusInternalServerError)
		return err
	}

	isM3U8 := strings.HasSuffix(parsedURL.Path, ".m3u8")

	// 如果启用了频道缓存且是 M3U8，使用频道缓存
	if config.ChannelCacheEnabled && isM3U8 && globalChannelCacheManager != nil {
		return ProxyM3U8WithChannelCache(c, originalURL, w, proxyAddress, config)
	}

	return ProxyStreamWithConfig(c, originalURL, w, proxyAddress, config)
}

// ProxyStreamWithConfig 使用指定配置代理流
func ProxyStreamWithConfig(c *http.Client, originalURL string, w http.ResponseWriter, proxyAddress string, config *CacheConfig) error {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		log.Printf("Error parsing originalURL %s: %v", originalURL, err)
		http.Error(w, "Failed to parse stream URL", http.StatusInternalServerError)
		return err
	}

	// 1. Create an HTTP request to the original upstream URL
	req, err := http.NewRequest("GET", originalURL, nil)
	if err != nil {
		log.Printf("Error creating request to upstream URL %s: %v", originalURL, err)
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 2. Check if the URL is an M3U8 playlist
	isM3U8 := strings.HasSuffix(parsedURL.Path, ".m3u8")

	if isM3U8 && config.EnableCache {
		// 使用带缓存的 M3U8 代理
		return proxyM3U8WithCache(c, originalURL, w, proxyAddress, config, parsedURL)
	} else if isM3U8 {
		// 使用不带缓存的 M3U8 代理（原有逻辑）
		return proxyM3U8WithoutCache(c, originalURL, w, proxyAddress, parsedURL)
	} else if config.EnableCache {
		// 使用带缓冲的非 M3U8 流代理
		return proxyStreamWithBuffer(c, originalURL, w, config)
	} else {
		// 使用不带缓冲的非 M3U8 流代理（原有逻辑）
		return proxyStreamWithoutBuffer(c, originalURL, w)
	}
}

// proxyM3U8WithCache 使用缓存代理 M3U8 流
func proxyM3U8WithCache(c *http.Client, originalURL string, w http.ResponseWriter, proxyAddress string, config *CacheConfig, parsedURL *url.URL) error {
	log.Printf("使用缓存代理 M3U8 播放列表：%s", originalURL)

	// 首先获取 M3U8 内容
	req, err := http.NewRequest("GET", originalURL, nil)
	if err != nil {
		log.Printf("Error creating request to M3U8 URL %s: %v", originalURL, err)
		http.Error(w, "Failed to create M3U8 request", http.StatusInternalServerError)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error fetching M3U8 from %s: %v", originalURL, err)
		http.Error(w, "Failed to fetch M3U8", http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading M3U8 body from %s: %v", originalURL, err)
		http.Error(w, "Failed to read M3U8 content", http.StatusInternalServerError)
		return err
	}

	// 解析 M3U8 内容，提取分片 URL
	bodyString := string(bodyBytes)
	baseURL := url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   parsedURL.Path[:strings.LastIndex(parsedURL.Path, "/")+1],
	}

	var segmentURLs []string
	for _, line := range strings.Split(bodyString, "\n") {
		cleanLine := strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(cleanLine, "#") || cleanLine == "" {
			continue
		}
		// 解析分片 URL
		segmentURL, err := url.Parse(cleanLine)
		if err != nil {
			continue
		}
		if !segmentURL.IsAbs() {
			segmentURL = baseURL.ResolveReference(segmentURL)
		}
		segmentURLs = append(segmentURLs, segmentURL.String())
	}

	if len(segmentURLs) == 0 {
		log.Printf("M3U8 中没有找到分片：%s", originalURL)
		http.Error(w, "No segments found in M3U8", http.StatusInternalServerError)
		return fmt.Errorf("no segments found")
	}

	log.Printf("M3U8 包含 %d 个分片，开始预取前 %d 个", len(segmentURLs), config.M3U8PreloadCount)

	// 创建缓存管理器
	cacheManager := NewM3U8CacheManager(config, c, segmentURLs, baseURL.String())
	cacheManager.StartPreload()

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(http.StatusOK)

	// 重新生成 M3U8 内容，使用代理 URL
	var newLines []string
	newLines = append(newLines, "#EXTM3U")

	// 添加原始 M3U8 中的元数据行
	for _, line := range strings.Split(bodyString, "\n") {
		cleanLine := strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(cleanLine, "#EXTM3U") {
			continue
		}
		if strings.HasPrefix(cleanLine, "#") {
			newLines = append(newLines, cleanLine)
		} else if strings.HasPrefix(cleanLine, "http") {
			// 这是分片 URL，替换为代理 URL
			encodedURL := url.QueryEscape(cleanLine)
			proxyURL := fmt.Sprintf("%s/proxy/stream?url=%s", proxyAddress, encodedURL)
			newLines = append(newLines, proxyURL)
		}
	}

	modifiedBody := strings.Join(newLines, "\n")
	_, err = w.Write([]byte(modifiedBody))
	if err != nil {
		log.Printf("Error writing M3U8: %v", err)
		cacheManager.Close()
		return err
	}

	// 注意：这里我们只返回重写后的 M3U8 播放列表
	// 实际的分片数据会在客户端请求代理 URL 时使用缓存
	cacheManager.Close()
	log.Printf("M3U8 播放列表已发送：%s", originalURL)
	return nil
}

// proxyM3U8WithoutCache 不使用缓存代理 M3U8（原有逻辑）
func proxyM3U8WithoutCache(c *http.Client, originalURL string, w http.ResponseWriter, proxyAddress string, parsedURL *url.URL) error {
	log.Printf("Rewriting M3U8 playlist without cache from %s", originalURL)

	req, err := http.NewRequest("GET", originalURL, nil)
	if err != nil {
		log.Printf("Error creating request to M3U8 URL %s: %v", originalURL, err)
		http.Error(w, "Failed to create M3U8 request", http.StatusInternalServerError)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error fetching M3U8 from %s: %v", originalURL, err)
		http.Error(w, "Failed to fetch M3U8", http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading M3U8 body from %s: %v", originalURL, err)
		http.Error(w, "Failed to read M3U8 content", http.StatusInternalServerError)
		return err
	}

	bodyString := string(bodyBytes)
	var newLines []string
	baseURL := url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
	}

	for _, line := range strings.Split(bodyString, "\n") {
		cleanLine := strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(cleanLine, "#") || cleanLine == "" {
			newLines = append(newLines, cleanLine)
			continue
		}
		segmentURL, err := url.Parse(cleanLine)
		if err != nil {
			log.Printf("Error parsing segment URL '%s' in M3U8 from %s: %v", cleanLine, originalURL, err)
			newLines = append(newLines, cleanLine)
			continue
		}
		if !segmentURL.IsAbs() {
			segmentURL = baseURL.ResolveReference(segmentURL)
		}
		encodedSegmentURL := url.QueryEscape(segmentURL.String())
		proxySegmentURL := fmt.Sprintf("%s/proxy/stream?url=%s", proxyAddress, encodedSegmentURL)
		newLines = append(newLines, proxySegmentURL)
	}

	modifiedBody := strings.Join(newLines, "\n")
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write([]byte(modifiedBody))
	if err != nil {
		log.Printf("Error writing modified M3U8 body for %s: %v", originalURL, err)
		return err
	}
	log.Printf("Finished rewriting and sending M3U8 from %s", originalURL)
	return nil
}

// proxyStreamWithBuffer 使用缓冲区代理非 M3U8 流
func proxyStreamWithBuffer(c *http.Client, originalURL string, w http.ResponseWriter, config *CacheConfig) error {
	log.Printf("使用缓冲区代理流：%s (缓冲大小：%d KB)", originalURL, config.StreamBufferSize/1024)

	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(config.PreloadTimeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 创建缓冲流
	bufferManager := NewStreamBufferManager(config, client)
	bufferedStream, err := bufferManager.CreateBufferedStream(originalURL)
	if err != nil {
		log.Printf("Error creating buffered stream from %s: %v", originalURL, err)
		http.Error(w, "Failed to create buffered stream", http.StatusBadGateway)
		return err
	}
	defer bufferManager.Close(bufferedStream)

	// 设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// 从缓冲区读取数据并发送给客户端
	buf := make([]byte, 8192)
	for {
		n, err := bufferedStream.Read(buf)
		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				return writeErr
			}
			// 刷新缓冲区，确保数据立即发送
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				log.Printf("Stream finished: %s", originalURL)
				return nil
			}
			log.Printf("Error reading from buffered stream %s: %v", originalURL, err)
			return err
		}
	}
}

// proxyStreamWithoutBuffer 不使用缓冲代理非 M3U8 流（原有逻辑）
func proxyStreamWithoutBuffer(c *http.Client, originalURL string, w http.ResponseWriter) error {
	log.Printf("Starting to proxy non-M3U8 stream without buffer from %s", originalURL)

	req, err := http.NewRequest("GET", originalURL, nil)
	if err != nil {
		log.Printf("Error creating request to upstream URL %s: %v", originalURL, err)
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error fetching stream from upstream URL %s: %v", originalURL, err)
		http.Error(w, "Failed to fetch stream from upstream", http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	// Copy relevant headers
	for key, val := range resp.Header {
		if strings.ToLower(key) == "connection" || strings.ToLower(key) == "transfer-encoding" || strings.ToLower(key) == "keep-alive" || strings.ToLower(key) == "proxy-authenticate" || strings.ToLower(key) == "proxy-authorization" || strings.ToLower(key) == "te" || strings.ToLower(key) == "trailers" || strings.ToLower(key) == "upgrade" {
			continue
		}
		w.Header()[key] = val
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil && err != io.EOF {
		log.Printf("Error while proxying non-M3U8 stream from %s: %v", originalURL, err)
		return err
	}
	log.Printf("Finished proxying non-M3U8 stream from %s", originalURL)
	return nil
}

// StreamCache 流缓存上下文，用于管理单个流的生命周期
type StreamCache struct {
	Config       *CacheConfig
	Client       *http.Client
	ProxyAddress string
	Ctx          context.Context
	Cancel       context.CancelFunc
}

// NewStreamCache 创建新的流缓存上下文
func NewStreamCache(config *CacheConfig, client *http.Client, proxyAddress string) *StreamCache {
	ctx, cancel := context.WithCancel(context.Background())
	return &StreamCache{
		Config:       config,
		Client:       client,
		ProxyAddress: proxyAddress,
		Ctx:          ctx,
		Cancel:       cancel,
	}
}

// Close 关闭流缓存
func (sc *StreamCache) Close() {
	sc.Cancel()
}
