package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type thuum struct {
	word         string
	ipa          string
	meanings     []string
	dragonScript string
	body         string
}

func (t thuum) mdxString(cssFile string) string {
	r := strings.NewReplacer("'", "")
	t.dragonScript = r.Replace(strings.ToUpper(t.word))
	r = strings.NewReplacer(
		"AA", "1",
		"AH", "4",
		"EI", "2",
		"EY", "9",
		"II", "3",
		"IR", "7",
		"OO", "8",
		"UU", "5",
		"UR", "6",
	)
	t.dragonScript = r.Replace(t.dragonScript)
	for i, meaning := range t.meanings {
		dotPos := strings.Index(meaning, ".") + 1
		if dotPos == 0 {
			t.meanings[i-1] = t.meanings[i-1] + " <i>" + meaning + "</i>"
			t.meanings = append(t.meanings[:i], t.meanings[i+1:]...)
		}
	}
	for _, meaning := range t.meanings {
		meaningStr := "<p><i>" + meaning + "</p>\r\n"
		dotPos := strings.Index(meaningStr, ".") + 1
		meaningStr = meaningStr[:dotPos] + "</i>" + meaningStr[dotPos:]
		t.body = t.body + meaningStr
	}
	t.body = strings.TrimSuffix(t.body, "\r\n")
	b, err := ioutil.ReadFile("template.txt")
	if err != nil {
		fmt.Println(err)
	}
	template := string(b)
	r = strings.NewReplacer(
		"{word}", t.word,
		"{css}", "\""+cssFile+"\"",
		"{dragon_script}", t.dragonScript,
		"{ipa}", t.ipa,
		"{body}", t.body,
	)
	return r.Replace(template)
}

func main() {
	pages := "ABDEFGHIJKLMNOPQRSTUVWYZ"
	thuumme := make(map[string]thuum)
	for _, v := range pages {
		page := string(v)
		fmt.Println(page + "...")
		resp, err := http.Get("https://www.thuum.org/dictionary.php?letter=" + page)
		if err != nil {
			fmt.Println(err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		htmlStr := strings.NewReader(string(b))
		doc, err := goquery.NewDocumentFromReader(htmlStr)
		resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
		doc.Find("div.dic-listing ").Each(func(i int, s *goquery.Selection) {
			meanings := make([]string, 0, 5)
			s.Find("div.info p").Contents().Each(func(i int, s *goquery.Selection) {
				if m := strings.TrimSpace(s.Text()); m != "" {
					meanings = append(meanings, m)
				}
			})
			ipa := strings.TrimSpace(s.Find("div.info").Contents().First().Text())
			word := s.Find("a").Text()
			if _, exist := thuumme[word]; exist {
				t := thuumme[word]
				t.meanings = append(t.meanings, meanings...)
				thuumme[word] = t
			} else {
				t := thuum{word: word, ipa: ipa, meanings: meanings}
				thuumme[word] = t
			}
		})
		time.Sleep(time.Second)
	}

	words := make([]string, 0)
	for k := range thuumme {
		words = append(words, k)
	}
	sort.Strings(words)
	for i, k := range words {
		f, _ := os.OpenFile("../src/thuum.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		defer f.Close()
		if i == len(words)-1 {
			f.WriteString(thuumme[k].mdxString("thuum.css"))
		} else {
			f.WriteString(thuumme[k].mdxString("thuum.css") + "\r\n")
		}
	}
}
