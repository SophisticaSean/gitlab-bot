package gitlab

import (
	"encoding/json"
	"fmt"
	"github.com/SophisticaSean/gitlab-bot/model"
)

func (gitlab Gitlab) searchProjects(name string) (projects []model.Project) {
	path := fmt.Sprintf("projects?search=%s", name)
	body := gitlab.Get(path)

	json.Unmarshal(body, &projects)
	return projects
}
