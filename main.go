package main

import (
	//"fmt"
	//"time"
  //"flag"
  "strings"

	"github.com/SophisticaSean/gitlab-bot/gitlab"
  "github.com/davecgh/go-spew/spew"
  "github.com/SophisticaSean/gitlab-bot/model"
)

var job_count int
var app_id string
var job_name string
var owner_name string

func main() {

	gl := gitlab.New()

  name := "juno"
  whitelisted_namespaces := []string{"backend", "platform", "frontend"}
  results := gl.SearchProjects(name)
  filtered := []model.Project{}
  for _, proj := range results {
    if proj.Name == name {
      for _, namespace := range whitelisted_namespaces {
        if strings.Contains(proj.PathWithNamespace, namespace) {
          filtered = append(filtered, proj)
        }
      }
    }
  }
  spew.Dump(filtered)
}
