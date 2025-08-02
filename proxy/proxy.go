package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// ProxyStream handles the core logic for proxying an HTTP stream.
// It takes the original upstream URL, fetches the stream, and pipes it to the client's response writer.
// It also takes the proxyAddress to rewrite M3U8 playlists if necessary.
func ProxyStream(c *http.Client, originalURL string, w http.ResponseWriter, proxyAddress string) error {
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

	// 2. Execute the request to the upstream server
	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error fetching stream from upstream URL %s: %v", originalURL, err)
		http.Error(w, "Failed to fetch stream from upstream", http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	// 3. Check if the response is an M3U8 playlist
	contentType := resp.Header.Get("Content-Type")
	isM3U8 := strings.HasSuffix(parsedURL.Path, ".m3u8") || strings.Contains(contentType, "application/vnd.apple.mpegurl") || strings.Contains(contentType, "application/x-mpegURL")

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
			// Clean up the line by stripping trailing carriage return (\r)
			cleanLine := strings.TrimSuffix(line, "\r")
			if strings.HasPrefix(cleanLine, "#") || cleanLine == "" {
				newLines = append(newLines, cleanLine)
				continue
			}
			// This is a media segment or another playlist URL
			segmentURL, err := url.Parse(cleanLine)
			if err != nil {
				log.Printf("Error parsing segment URL '%s' in M3U8 from %s: %v", cleanLine, originalURL, err)
				newLines = append(newLines, cleanLine) // Keep original line if parsing fails
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
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl") // Standard M3U8 content type
		w.WriteHeader(resp.StatusCode)
		_, err = w.Write([]byte(modifiedBody))
		if err != nil {
			log.Printf("Error writing modified M3U8 body for %s: %v", originalURL, err)
			return err
		}
		log.Printf("Finished rewriting and sending M3U8 from %s", originalURL)
	} else {
		// 4. If not M3U8, stream the response body directly
		log.Printf("Starting to proxy non-M3U8 stream from %s", originalURL)
		// Copy relevant headers
		for key, val := range resp.Header {
			// Avoid copying hop-by-hop headers
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
	}

	return nil
}
