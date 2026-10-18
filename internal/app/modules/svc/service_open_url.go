package svc

import (
	"strings"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
)

func (s *Service) OpenURL(browser, url string) error {
	proxy := psl.GetConfig().Settings.VPN
	return platform.Get().OpenBrowser(browser, url, proxy, false)
}

func (s *Service) OpenURLAsApp(browser, url string) error {
	proxy := psl.GetConfig().Settings.VPN
	return platform.Get().OpenBrowser(browser, url, proxy, true)
}

func (s *Service) SearchBooksOnline(query string) error {
	q := strings.ReplaceAll(query, " ", "+")
	urls := []string{
		"https://openlibrary.org/search?q=" + q,
		"https://z-lib.id/s?q=" + q,
	}
	for _, u := range urls {
		if err := s.OpenURL("chrome", u); err != nil {
			s.logger.WithError(err).WithField("url", u).Warn("open books url failed")
		}
	}
	return nil
}

func (s *Service) SearchVideosOnline(query string) error {
	q := strings.ReplaceAll(query, " ", "+")
	urls := []string{
		"https://search.bilibili.com/all?keyword=" + q,
		"https://www.youtube.com/results?search_query=" + q,
	}
	for _, u := range urls {
		if err := s.OpenURL("chrome", u); err != nil {
			s.logger.WithError(err).WithField("url", u).Warn("open videos url failed")
		}
	}
	return nil
}
