package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	linkCheckRequestTimeout = 10 * time.Second
	linkCheckGlobalDelay    = 750 * time.Millisecond
	linkCheckDomainDelay    = 4 * time.Second
)

type clubLink struct {
	club      *Club
	linkType  string
	linkValue string
	host      string
}

func CheckLinks(data *Data) error {
	client := &http.Client{Timeout: linkCheckRequestTimeout}
	lastRequestAt := make(map[string]time.Time)
	lastGlobalRequestAt := time.Time{}
	links := collectClubLinks(data.Clubs)

	if len(links) == 0 {
		fmt.Println("checking links: no club links found")
		return nil
	}

	fmt.Printf("checking links: %d links\n", len(links))

	for i, link := range links {
		if i == 0 || i%10 == 0 {
			fmt.Printf("checking links: %d/%d\n", i+1, len(links))
		}

		waitForLinkCheckSlot(lastGlobalRequestAt, lastRequestAt[link.host])

		broken, err := isBrokenClubLink(client, link.linkValue)
		now := time.Now()
		lastGlobalRequestAt = now
		lastRequestAt[link.host] = now

		if err != nil {
			fmt.Printf("broken %s link for %q (%s): %s (%v)\n", link.linkType, link.club.Name, link.club.City.Name, link.linkValue, err)
			continue
		}
		if broken {
			fmt.Printf("broken %s link for %q (%s): %s\n", link.linkType, link.club.Name, link.club.City.Name, link.linkValue)
		}
	}

	fmt.Printf("checking links: done (%d/%d)\n", len(links), len(links))

	return nil
}

func collectClubLinks(clubs []*Club) []clubLink {
	links := make([]clubLink, 0, len(clubs)*5)
	for _, club := range clubs {
		links = appendClubLink(links, club, "instagram", club.Instagram)
		links = appendClubLink(links, club, "strava", club.StravaClub)
		links = appendClubLink(links, club, "whatsapp", club.Whatsapp)
		links = appendClubLink(links, club, "signal", club.Signal)
		links = appendClubLink(links, club, "website", club.Website)
	}
	return links
}

func appendClubLink(links []clubLink, club *Club, linkType, linkValue string) []clubLink {
	linkValue = strings.TrimSpace(linkValue)
	if linkValue == "" {
		return links
	}

	parsedURL, err := url.Parse(linkValue)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return append(links, clubLink{club: club, linkType: linkType, linkValue: linkValue})
	}

	return append(links, clubLink{
		club:      club,
		linkType:  linkType,
		linkValue: linkValue,
		host:      normalizeLinkHost(parsedURL.Hostname()),
	})
}

func normalizeLinkHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimPrefix(host, "www.")
	return host
}

func waitForLinkCheckSlot(lastGlobalRequestAt, lastHostRequestAt time.Time) {
	now := time.Now()
	if !lastGlobalRequestAt.IsZero() {
		if wait := lastGlobalRequestAt.Add(linkCheckGlobalDelay).Sub(now); wait > 0 {
			time.Sleep(wait)
			now = time.Now()
		}
	}
	if !lastHostRequestAt.IsZero() {
		if wait := lastHostRequestAt.Add(linkCheckDomainDelay).Sub(now); wait > 0 {
			time.Sleep(wait)
		}
	}
}

func isBrokenClubLink(client *http.Client, rawURL string) (bool, error) {
	requestContext, cancel := context.WithTimeout(context.Background(), linkCheckRequestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, rawURL, nil)
	if err != nil {
		return true, err
	}
	request.Header.Set("Range", "bytes=0-0")
	request.Header.Set("User-Agent", "socialrunclubs.de link-checker/1.0 (+https://socialrunclubs.de)")

	response, err := client.Do(request)
	if err != nil {
		return true, err
	}
	defer response.Body.Close()

	return response.StatusCode >= http.StatusBadRequest, nil
}
