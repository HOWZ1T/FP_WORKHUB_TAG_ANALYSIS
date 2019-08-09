package main

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const XhrUrl = "https://functional.works-hub.com/api/graphql?query=query%20jobs_search(%24vertical%3Avertical%2C%24search_term%3AString!%2C%24preset_search%3AString%2C%24page%3AInt!%2C%24filters%3ASearchFiltersInput)%7Bjobs_search(vertical%3A%24vertical%2Csearch_term%3A%24search_term%2Cpreset_search%3A%24preset_search%2Cpage%3A%24page%2Cfilters%3A%24filters)%7BnumberOfPages%2CnumberOfHits%2ChitsPerPage%2Cpage%2Cfacets%7Battr%2Cvalue%2Ccount%7D%2CsearchParams%7Blabel%2Cquery%2Cfilters%7Bremote%2CroleType%2CsponsorshipOffered%2Cpublished%2Ctags%2Cmanager%2Clocation%7Bcities%2CcountryCodes%2Cregions%7D%2Cremuneration%7Bmin%2Cmax%2Ccurrency%2CtimePeriod%7D%7D%7D%2Cpromoted%7B...jobCardFields%7D%2Cjobs%7B...jobCardFields%7D%7D%2Ccity_info%7Bcity%2Ccountry%2CcountryCode%2Cregion%7D%2Cremuneration_ranges%7Bcurrency%2CtimePeriod%2Cmin%2Cmax%7D%7D%20fragment%20jobCardFields%20on%20JobCard%7Bid%2Cslug%2Ctitle%2CcompanyName%2Ctagline%2Clocation%7Bcity%2Cstate%2Ccountry%2CcountryCode%7D%2Cremuneration%7Bcompetitive%2Ccurrency%2CtimePeriod%2Cmin%2Cmax%2Cequity%7D%2Clogo%2Ctags%2Cpublished%2CuserScore%2CroleType%2CsponsorshipOffered%2Cremote%2CcompanyId%7D&variables=%7B%22search_term%22%3A%22%22%2C%22preset_search%22%3Anull%2C%22page%22%3A8%2C%22filters%22%3A%7B%7D%2C%22vertical%22%3A%22functional%22%7D"

func closeResp(resp *http.Response) {
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
}

func makeHttpRequest(uri string) (resp *http.Response, err error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err = client.Get(uri)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getJobTags() []string {
	resp, err := makeHttpRequest(XhrUrl)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer closeResp(resp)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	jsonRaw := buf.String()
	promoted := gjson.Get(jsonRaw, "data.jobs_search.promoted.#.tags")
	jobs := gjson.Get(jsonRaw, "data.jobs_search.jobs.#.tags")
	var tags []string
	// notice: we want to have duplicates in tags for counting the popularity of the tags

	for _, items := range promoted.Array() {
		for _, tag := range items.Array() {
			tags = append(tags, strings.ToLower(tag.String()))
		}
	}

	for _, items := range jobs.Array() {
		for _, tag := range items.Array() {
			tags = append(tags, strings.ToLower(tag.String()))
		}
	}

	return tags
}

func analyzeTags(tags []string) {
	m := make(map[string]int)
	for _, tag := range tags {
		// ok bool is true if key exists and false if key doesn't exist
		if _, ok := m[tag]; ok {
			m[tag] += 1
		} else {
			// adding new key to map
			m[tag] = 1
		}
	}

	// sorting map by value
	// defining entry struct
	type entry struct {
		Key   string
		Value int
	}

	// creating slice of entry structs
	var ss []entry
	for k, v := range m {
		ss = append(ss, entry{k, v})
	}

	// sorting the slice by value
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for _, kv := range ss {
		fmt.Printf("%30s : %-3d\n", kv.Key, kv.Value)
	}
}

func main() {
	tags := getJobTags()
	analyzeTags(tags)
}
