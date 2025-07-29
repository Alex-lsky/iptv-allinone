package proxy

import (
	"io"
	"log"
	"net/http"
)

// ProxyStream handles the core logic for proxying an HTTP stream.
// It takes the original upstream URL, fetches the stream, and pipes it to the client's response writer.
func ProxyStream(c *http.Client, originalURL string, w http.ResponseWriter) error {
	// 1. Create an HTTP request to the original upstream URL
	req, err := http.NewRequest("GET", originalURL, nil)
	if err != nil {
		log.Printf("Error creating request to upstream URL %s: %v", originalURL, err)
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return err
	}

	// 2. (Optional) Copy relevant headers from the client's incoming request
	// to the upstream request, such as User-Agent, Range, etc.
	// For now, we'll keep it simple and not forward headers.

	// 3. Execute the request to the upstream server
	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error fetching stream from upstream URL %s: %v", originalURL, err)
		// It's possible the upstream is temporarily unavailable.
		// A 502 Bad Gateway is a standard response for this scenario.
		http.Error(w, "Failed to fetch stream from upstream", http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	// 4. Copy relevant headers from the upstream response to our response writer
	// This is important for things like Content-Type, Content-Length, etc.
	// We should be selective about which headers to copy to avoid issues.
	// Common headers to forward:
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	// Add more headers as needed, but be cautious.
	// Avoid copying hop-by-hop headers like Connection, Transfer-Encoding, etc.

	// 5. Set the status code to be the same as the upstream response
	w.WriteHeader(resp.StatusCode)

	// 6. Stream the response body from the upstream server directly to the client
	// io.Copy is efficient for streaming as it doesn't load the entire body into memory.
	log.Printf("Starting to proxy stream from %s", originalURL)
	_, err = io.Copy(w, resp.Body)
	if err != nil && err != io.EOF {
		// Log the error, but it might not be possible to send an error response
		// to the client if the stream has already started.
		log.Printf("Error while proxying stream from %s: %v", originalURL, err)
		// Depending on how much data has been sent, sending an error might not be effective.
		// The connection might already be closed or corrupted.
		return err
	}
	log.Printf("Finished proxying stream from %s", originalURL)

	return nil
}
