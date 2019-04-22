package gitlab

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type Gitlab struct {
	BaseURL string
	Token   string
}

func New() Gitlab {
	// set token
	token := os.Getenv("gitlab_token")
	if token == "" {
		panic("gitlab_token env var not set")
	}
	// set base_url
	base_url := os.Getenv("gitlab_base_url")
	if base_url == "" {
		panic("gitlab_base_url env var not set (remove any trailing slashes)")
	}

	return Gitlab{BaseURL: base_url, Token: token}
}

func (gitlab Gitlab) PostNoData(path string) (body []byte) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v4/%s", gitlab.BaseURL, path)
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Add("PRIVATE-TOKEN", gitlab.Token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		spew.Dump(body)
		spew.Dump(url)
		panic(fmt.Sprintf("StatusCode not 200 or 201: %d", resp.StatusCode))
	}

	return body
}

func (gitlab Gitlab) Get(path string) (body []byte) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v4/%s", gitlab.BaseURL, path)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", gitlab.Token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	if resp.StatusCode != 200 {
		spew.Dump(body)
		panic(fmt.Sprintf("StatusCode not 200: %d", resp.StatusCode))
	}

	return body
}
