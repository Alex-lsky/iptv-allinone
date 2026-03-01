package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// ProxyStream handles the core logic for proxying an HTTP stream.
func ProxyStream(c *http.Client, originalURL string, w http.ResponseWriter, proxyAddress string) error {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		log.Printf("Error parsing originalURL %s: %v", originalURL, err)
		http.Error(w, "Failed to parse stream URL", http.StatusInternalServerError)
		return err
	}

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

	isM3U8 := strings.HasSuffix(parsedURL.Path, ".m3u8")

	if isM3U8 {
		log.Printf("Rewriting M3U8 playlist from %s", originalURL)
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
			proxySegmentURL := strings.Replace(proxyAddress, "passwall.lhtsky.top", "localhost", 1) + "/proxy/stream?url=" + encodedSegmentURL
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
	} else {
		log.Printf("Starting to proxy non-M3U8 stream from %s", originalURL)
		for key, val := range resp.Header {
			if strings.ToLower(key) == "connection" || strings.ToLower(key) == "transfer-encoding" || strings.ToLower(key) == "keep-alive" {
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
	}

	return nil
}
