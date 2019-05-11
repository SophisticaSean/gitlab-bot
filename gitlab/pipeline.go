package gitlab

import (
	"encoding/json"
	"fmt"
	"github.com/SophisticaSean/gitlab-bot/model"
)

func (gitlab Gitlab) GetPipeline(project_id, pipeline_id string) (pipeline model.Pipeline) {
	path := fmt.Sprintf("projects/%s/pipelines/%s", project_id, pipeline_id)
	body := gitlab.Get(path)

	json.Unmarshal(body, &pipeline)
	return pipeline
}

func (gitlab Gitlab) TriggerPipeline(ref_branch, project_id string, vars map[string]string) (pipeline model.Pipeline) {
  stringVars := formatVars(vars)
	path := fmt.Sprintf("projects/%s/pipeline?ref=%s", project_id, ref_branch)
  if stringVars != "" {
    path = path + stringVars
  }
	body := gitlab.PostNoData(path)

	json.Unmarshal(body, &pipeline)
	return pipeline
}

func formatVars(vars map[string]string) string {
  out := ""
  for k, v := range vars {
    out = out + fmt.Sprintf("&variables[][key]=%s&variables[][value]=%s", k, v)
  }
  return out
}
