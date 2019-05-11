package model

import "time"

type Pipeline struct {
	ID         int      `json:"id"`
	Sha        string      `json:"sha"`
	Ref        string      `json:"ref"`
	Status     string      `json:"status"`
	BeforeSha  string      `json:"before_sha"`
	Tag        bool        `json:"tag"`
	YamlErrors interface{} `json:"yaml_errors"`
	User       struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"user"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	StartedAt   interface{} `json:"started_at"`
	FinishedAt  interface{} `json:"finished_at"`
	CommittedAt interface{} `json:"committed_at"`
	Duration    interface{} `json:"duration"`
	Coverage    interface{} `json:"coverage"`
	WebURL      string      `json:"web_url"`
}

type PipelineVars []struct {
	Key          string `json:"key"`
	VariableType string `json:"variable_type,omitempty"`
	Value        string `json:"value"`
}
