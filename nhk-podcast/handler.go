package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/mmcdole/gofeed"

	"gocloud.dev/blob"

	//use gcs
	_ "gocloud.dev/blob/gcsblob"
)

const (
	GCP_BUCKET   = "gs://laputa-public"
	GCP_OBJECT   = "nhk.xml"
	NHK_FEED_URL = "https://www.nhk.or.jp/s-media/news/podcast/list/v1/all.xml"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(NHK_FEED_URL)
	if err != nil {
		log.Printf("failed to get and parse nhk rss feed, %s", err)
		sendError(w, http.StatusInternalServerError)
		return
	}

	keepItems := filter(feed.Items)
	feed.Items = keepItems

	ctx := context.Background()

	b, err := blob.OpenBucket(ctx, GCP_BUCKET)
	if err != nil {
		log.Printf("Failed to setup bucket: %s", err)
		sendError(w, http.StatusInternalServerError)
		return
	}
	defer b.Close()

	rssFileWriter, err := b.NewWriter(ctx, GCP_OBJECT, &blob.WriterOptions{
		ContentType: "application/rss+xml;charset=utf-8",
	})
	if err != nil {
		log.Printf("Failed to write rss file to bucket: %s", err)
		sendError(w, http.StatusInternalServerError)
		return
	}

	err = writeRss(rssFileWriter, feed)
	if err != nil {
		log.Printf("failed to generate final rss %s", err)
		sendError(w, http.StatusInternalServerError)
		return
	}

	err = rssFileWriter.Close()
	if err != nil {
		log.Printf("failed to write rss file: %s", err)
		sendError(w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sendError(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

func filter(items []*gofeed.Item) []*gofeed.Item {
	containsAny := func(s string, keys []string) bool {
		for _, k := range keys {
			if strings.Contains(s, k) {
				return true
			}
		}

		return false
	}

	titleFilterKeys := []string{"午前７時", "夜７時"}

	keepItems := []*gofeed.Item{}

	for _, item := range items {
		if containsAny(item.Title, titleFilterKeys) {
			log.Printf("found item to keep: %#v", item)
			keepItems = append(keepItems, item)
		}
	}

	return keepItems
}

func writeRss(wr io.Writer, feed *gofeed.Feed) error {
	t, err := template.ParseFiles("./nhk-feed.template")
	if err != nil {
		return err
	}

	t.Execute(wr, feed)

	return nil
}
