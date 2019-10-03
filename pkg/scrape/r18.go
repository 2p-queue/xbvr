package scrape

import (
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/nleeper/goment"
	"github.com/thoas/go-funk"
)

func ScrapeR18(knownScenes []string, out *[]ScrapedScene, queryString string) error {
	siteCollector := colly.NewCollector(
		colly.AllowedDomains("www.r18.com"),
		// colly.CacheDir(siteCacheDir),
		colly.UserAgent(userAgent),
	)

	sceneCollector := colly.NewCollector(
		colly.AllowedDomains("www.r18.com"),
		colly.CacheDir(sceneCacheDir),
		colly.UserAgent(userAgent),
	)

	siteCollector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	sceneCollector.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	sceneCollector.OnHTML(`html`, func(e *colly.HTMLElement) {
		sc := ScrapedScene{}
		sc.SceneType = "VR"
		sc.Studio = "JAVR"
		sc.HomepageURL = strings.Split(e.Request.URL.String(), "?")[0]

		// Title
		e.ForEach(`h1 cite[itemprop=name]`, func(id int, e *colly.HTMLElement) {
			sc.Title = strings.TrimSpace(strings.Replace(e.Text, "[VR]", "", -1))
		})

		// Covers
		e.ForEach(`div.detail-single-picture img`, func(id int, e *colly.HTMLElement) {
			sc.Covers = append(sc.Covers, e.Attr("src"))
		})

		// Gallery
		e.ForEach(`section#product-gallery img.lazyOwl`, func(id int, e *colly.HTMLElement) {
			sc.Gallery = append(sc.Gallery, e.Attr("data-src"))
		})

		// Cast
		e.ForEach(`div[itemprop=actors] a`, func(id int, e *colly.HTMLElement) {
			sc.Cast = append(sc.Cast, strings.TrimSpace(e.Text))
		})

		// Tags
		e.ForEach(`div.pop-list a`, func(id int, e *colly.HTMLElement) {
			sc.Tags = append(sc.Tags, strings.TrimSpace(e.Text))
		})

		// Release date / Duration / Site
		e.ForEach(`div.product-details dd`, func(id int, e *colly.HTMLElement) {
			if id == 0 {
				tmpDate, _ := goment.New(strings.TrimSpace(e.Text), "MMM. DD, YYYY")
				sc.Released = tmpDate.Format("YYYY-MM-DD")
			}
			if id == 1 {
				tmpDuration, err := strconv.Atoi(strings.TrimSpace(strings.Replace(e.Text, "min.", "", -1)))
				if err == nil {
					sc.Duration = tmpDuration
				}
			}
			if id == 4 {
				sc.Site = strings.TrimSpace(e.Text)
			}
		})

		// Scene ID
		e.ForEach(`div.product-details dt:contains("DVD ID")+dd`, func(id int, e *colly.HTMLElement) {
			sc.SceneID = strings.TrimSpace(e.Text)
			sc.SiteID = strings.TrimSpace(e.Text)
		})

		sc.Tags = append(sc.Tags, "JAVR")
		*out = append(*out, sc)
	})

	siteCollector.OnHTML(`html`, func(e *colly.HTMLElement) {
		sceneURL := ""

		e.ForEach(`ul.cmn-list-product01 li a`, func(id int, e *colly.HTMLElement) {
			if id == 0 {
				sceneURL = strings.Split(e.Attr("href"), "?")[0]
			} else {
				sceneURL = ""
			}
		})

		// If scene exist in database, there's no need to scrape
		if !funk.ContainsString(knownScenes, sceneURL) {
			sceneCollector.Visit(sceneURL)
		}
	})

	if strings.Contains(queryString, "/movies/detail/") {
		return sceneCollector.Visit(queryString)
	} else {
		return siteCollector.Visit("https://www.r18.com/common/search/searchword=" + queryString + "/")
	}
}