package rssbot

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var feedXML = `
<?xml version="1.0" encoding="utf-8"?><?xml-stylesheet type="text/xsl" href="https://www.example.com/xsl"?>
<rss xmlns:media="http://search.yahoo.com/mrss/" version="2.0">
    <channel>
        <title>news | TEST</title>
        <description>newsdescription</description>
        <link>http://www.example.com</link>
        <language>en</language>
        <copyright>null</copyright>
        <webMaster>web@example.com</webMaster>
        <pubDate>Sun, 16 Aug 2020 13:43:44 +0300</pubDate>
        <lastBuildDate>Sun, 16 Aug 2020 13:30:00 +0300</lastBuildDate>
        <generator>Test generator</generator>
        <item>
            <title><![CDATA[Example news title1]]></title>
            <link>https://example.com/news/item1</link>
            <description><![CDATA[Example news description1 ]]></description>
            <media:thumbnail url='https://example.com/news/item1/picture1.jpg' height='75' width='75' />
            <guid isPermaLink="true">https://example.com/11111111</guid>
            <pubDate>Sun, 16 Aug 2020 13:30:00 +0300</pubDate>
            <category><![CDATA[Category1]]></category>
        </item>
        <item>
            <title><![CDATA[Example news title2]]></title>
            <link>https://example.com/news/item2</link>
            <description><![CDATA[Example news description2 ]]></description>
            <media:thumbnail url='https://example.com/news/item2/picture2.jpg' height='75' width='75' />
            <guid isPermaLink="true">https://example.com/222222222</guid>
            <pubDate>Sun, 16 Aug 2020 13:30:00 +0300</pubDate>
            <category><![CDATA[Category2]]></category>
		</item>
		<item>
			<title><![CDATA[Example news title3]]></title>
			<link>https://example.com/news/item3</link>
			<description><![CDATA[Example news description3 ]]></description>
			<media:thumbnail url='https://example.com/news/item3/picture3.jpg' height='75' width='75' />
			<guid isPermaLink="true">https://example.com/333333333</guid>
			<pubDate>Sun, 16 Aug 2020 13:30:00 +0300</pubDate>
			<category><![CDATA[Category3]]></category>
		</item>
		</channel>
		</rss>`

func TestRSSFeedFetchLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(feedXML))
	}))
	defer server.Close()
	n := RSSFeed{URL: server.URL}
	n.fetchLinks()
	if len(n.links) != 3 {
		t.Errorf("Wrong number of links extracted")
	}
	if n.links[1] != "https://example.com/news/item2" {
		t.Errorf("Unexpected link in position 2: %v", n.links[1])
	}
}

func TestRSSFeedReverse(t *testing.T) {
	links := []string{"example.com/1", "example.com/2", "example.com/3"}
	want := []string{"example.com/3", "example.com/2", "example.com/1"}
	n := RSSFeed{links: links}
	n.reverse()
	if !reflect.DeepEqual(want, n.links) {
		t.Errorf("Reversing the slice produced unexpected results, want:%v got:%v", want, n.links)
	}
}

func TestRSSFeedRemoveOlderThan(t *testing.T) {
	links := []string{"example.com/1", "example.com/2", "example.com/3"}
	want := []string{"example.com/1", "example.com/2"}
	olderThan := "example.com/3"
	n := RSSFeed{links: links}
	n.removeOlderThan(olderThan)
	if !reflect.DeepEqual(want, n.links) {
		t.Errorf("Removing older than %v failed, expected: %v, got: %v", olderThan, want, n.links)
	}
}

func TestTelegramAPICall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(feedXML))
	}))
	defer server.Close()
	chatID := "@test_chat"
	apiToken := "123456-aaaaaaa"

	api := TelegramAPI{apiToken: apiToken, apiURL: server.URL, chatID: chatID}
	params := struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: chatID,
		Text:   "test",
	}

	api.call("testCommand", params, )
}