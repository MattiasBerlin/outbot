package handlers

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"net/http"
)

type module struct {
	name   string
	levels []level
}

type level struct {
	level  int
	values []value
}

type value struct {
	name  string
	value int
}

func getModules() ([]module, error) {
	resp, err := http.Get("https://hades-star.fandom.com/wiki/Support_Modules")
	if err != nil {
		return nil, errors.Wrap(err, "unable to reach wiki")
	}
	defer resp.Body.Close()

	for {
		z := html.NewTokenizer(resp.Body)
		token := z.Next()
		fmt.Println(token.String())
		switch token {
		case html.ErrorToken:
			fmt.Println("Error:", z.Err())
			return nil, nil
		case html.StartTagToken, html.EndTagToken:
			tn, _ := z.TagName()
			fmt.Println("TagName:", tn)
		}
	}

	return nil, nil
}
