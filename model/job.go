package model

import (
	"strings"
	"sync"
	"time"
)

type JobSlice struct {
	Slice Jobs
	mux   sync.Mutex
}

func (js *JobSlice) Lock() {
	js.mux.Lock()
}

func (js *JobSlice) Unlock() {
	js.mux.Unlock()
}

type Jobs []Job

func (jobs Jobs) Combine(newJobs Jobs) Jobs {
	for _, job := range newJobs {
		job_id_in_list := false
		for _, upper_job := range jobs {
			if upper_job.ID == job.ID {
				job_id_in_list = true
			}
		}
		if !job_id_in_list {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

func (jobs Jobs) FilterByOwnerName(name string) Jobs {
	outjobs := Jobs{}
	for _, job := range jobs {
		if strings.Contains(job.User.Name, name) {
			outjobs = append(outjobs, job)
		}
	}
	return outjobs
}

func (jobs Jobs) FilterByJobName(name string) Jobs {
	outjobs := Jobs{}
	for _, job := range jobs {
		if strings.Contains(job.Name, name) {
			outjobs = append(outjobs, job)
		}
	}
	return outjobs
}

func (jobs Jobs) FilterByStatus(name string) Jobs {
	outjobs := Jobs{}
	for _, job := range jobs {
		if strings.Contains(job.Status, name) {
			outjobs = append(outjobs, job)
		}
	}
	return outjobs
}

func (jobs Jobs) FilterOutStatus(name string) Jobs {
	outjobs := Jobs{}
	for _, job := range jobs {
		if !strings.Contains(job.Status, name) {
			outjobs = append(outjobs, job)
		}
	}
	return outjobs
}

type Job struct {
	Commit struct {
		AuthorEmail string    `json:"author_email"`
		AuthorName  string    `json:"author_name"`
		CreatedAt   time.Time `json:"created_at"`
		ID          string    `json:"id"`
		Message     string    `json:"message"`
		ShortID     string    `json:"short_id"`
		Title       string    `json:"title"`
	} `json:"commit"`
	Coverage          float64   `json:"coverage"`
	CreatedAt         time.Time `json:"created_at"`
	StartedAt         time.Time `json:"started_at"`
	FinishedAt        time.Time `json:"finished_at"`
	Duration          float64   `json:"duration"`
	ArtifactsExpireAt time.Time `json:"artifacts_expire_at"`
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Pipeline          struct {
		ID     int    `json:"id"`
		Ref    string `json:"ref"`
		Sha    string `json:"sha"`
		Status string `json:"status"`
	} `json:"pipeline"`
	Ref       string        `json:"ref"`
	Artifacts []interface{} `json:"artifacts"`
	Runner    interface{}   `json:"runner"`
	Stage     string        `json:"stage"`
	Status    string        `json:"status"`
	Tag       bool          `json:"tag"`
	WebURL    string        `json:"web_url"`
	User      struct {
		ID           int         `json:"id"`
		Name         string      `json:"name"`
		Username     string      `json:"username"`
		State        string      `json:"state"`
		AvatarURL    string      `json:"avatar_url"`
		WebURL       string      `json:"web_url"`
		CreatedAt    time.Time   `json:"created_at"`
		Bio          interface{} `json:"bio"`
		Location     interface{} `json:"location"`
		PublicEmail  string      `json:"public_email"`
		Skype        string      `json:"skype"`
		Linkedin     string      `json:"linkedin"`
		Twitter      string      `json:"twitter"`
		WebsiteURL   string      `json:"website_url"`
		Organization string      `json:"organization"`
	} `json:"user"`
	ArtifactsFile struct {
		Filename string `json:"filename"`
		Size     int    `json:"size"`
	} `json:"artifacts_file,omitempty"`
}
