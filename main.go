package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/SophisticaSean/gitlab-bot/model"
	"github.com/davecgh/go-spew/spew"
)

var token string
var base_url string
var app_id = "21"
var job_count = 500

type JobSlice struct {
	Slice model.Jobs
	mux   sync.Mutex
}

func main() {
	// set token
	token = os.Getenv("gitlab_token")
	if token == "" {
		panic("gitlab_token env var not set")
	}
	// set base_url
	base_url = os.Getenv("gitlab_base_url")
	if base_url == "" {
		panic("gitlab_base_url env var not set (remove any trailing slashes)")
	}

	var jobs model.Jobs

	// wait for a new Unit Tests job
	for {
		fmt.Println("waiting for a running/pending new Unit Tests job")
		jobs = getJobsPageCount(app_id, 5)
		jobs = jobs.FilterByOwnerName("Sean")
		jobs = jobs.FilterByJobName("Unit Tests + Coverage")
		pending_jobs := jobs.FilterByStatus("pending")
		running_jobs := jobs.FilterByStatus("running")
		jobs = pending_jobs.Combine(running_jobs)
		if len(jobs) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// cancel all running jobs
	cancelJobs(app_id, jobs)

	// wait for a new canceled Unit Tests job
	for {
		fmt.Println("waiting for a new canceled Unit Tests job")
		jobs = getJobsPageCount(app_id, 5)
		jobs = jobs.FilterByOwnerName("Sean")
		jobs = jobs.FilterByJobName("Unit Tests + Coverage")
		jobs = jobs.FilterByStatus("canceled")
		if len(jobs) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	job := jobs[0]
	fmt.Printf("Using job_id: %d for new retry jobs\n", job.ID)
	time.Sleep(5 * time.Second)

	// clear out jobs
	jobs = model.Jobs{}
	failed := model.Jobs{}
	success := model.Jobs{}
	canceled := model.Jobs{}

	jobs = chunkRetryJobID(app_id, job.ID, job_count)
	if len(jobs) != job_count {
		panic(fmt.Sprintf("jobs is not %d!", job_count))
	}

	// wait for the jobs to complete
	for {
		current_jobs := model.Jobs{}
		for _, job := range jobs {
			job = getJob(app_id, job.ID)
			current_jobs = append(current_jobs, job)
			// cancel any jobs running longer than 10 minutes
			if job.Duration >= 600.0 {
				cancelJobs(app_id, model.Jobs{job})
			}
		}
		// filter by statuses
		success = current_jobs.FilterByStatus("success")
		fmt.Printf("Successful Jobs: %d\n", len(success))

		failed = current_jobs.FilterByStatus("failed")
		fmt.Printf("Failed Jobs: %d\n", len(failed))

		running := current_jobs.FilterByStatus("running")
		fmt.Printf("Running Jobs: %d\n", len(running))

		pending := current_jobs.FilterByStatus("pending")
		fmt.Printf("Pending Jobs: %d\n", len(pending))

		canceled = current_jobs.FilterByStatus("canceled")
		fmt.Printf("Canceled Jobs: %d\n", len(canceled))

		fmt.Println("")

		if (len(success) + len(failed) + len(canceled)) == len(jobs) {
			jobs = current_jobs
			break
		}
	}

	// compute and print the average amount of time our builds take to finish
	total_duration := 0.0
	for _, job := range jobs {
		total_duration = total_duration + job.Duration
	}
	avg_duration := total_duration / (float64(len(jobs)))
	fmt.Printf("Average duration seconds: %f\n", avg_duration)
	fmt.Printf("Average duration minutes: %f\n", avg_duration/float64(60))

	// compute and print our failure rate
	percentage_failed := (float64(len(failed)) / (float64(len(jobs) - len(canceled))))
	fmt.Printf("Failure rate: %f\n", percentage_failed*100)
	fmt.Println("Failed Jobs: ")

	for _, job := range failed {
		fmt.Printf("%s/backend/juno/-/jobs/%d\n", base_url, job.ID)
	}
	fmt.Println("")
}

func chunkRetryJobID(project_id string, job_id int, count int) model.Jobs {
	jobs := model.Jobs{}
	chunk_by := 25
	chunk_count := count / chunk_by
	remainder := count % chunk_by
	// retry jobs in chunks of 50
	for i := 0; i < chunk_count; i++ {
		new_jobs := retryJobID(project_id, job_id, chunk_by)
		jobs = jobs.Combine(new_jobs)
		time.Sleep(3 * time.Second)
	}
	// handle remainder
	if remainder > 0 {
		new_jobs := retryJobID(project_id, job_id, remainder)
		jobs = jobs.Combine(new_jobs)
	}
	return jobs
}

// retry job_id in project_id count times
//  return the array of new jobs
func retryJobID(project_id string, job_id int, count int) model.Jobs {
	out_wg := sync.WaitGroup{}
	job_slice := JobSlice{}
	for i := 0; i < count; i++ {
		out_wg.Add(1)
		go retryJobAsync(project_id, job_id, &out_wg, &job_slice)
	}
	out_wg.Wait()
	return job_slice.Slice
}

func retryJobAsync(project_id string, job_id int, new_wg *sync.WaitGroup, js *JobSlice) {
	defer new_wg.Done()
	defer js.mux.Unlock()
	new_job := retryJob(project_id, job_id)

	// lock the mutex and append our new job
	js.mux.Lock()
	js.Slice = append(js.Slice, new_job)
}

func cancelJobs(project_id string, jobs model.Jobs) {
	out_wg := sync.WaitGroup{}
	for _, j := range jobs {
		out_wg.Add(1)
		go func(job model.Job, wg *sync.WaitGroup) {
			defer wg.Done()
			cancelJob(project_id, job.ID)
		}(j, &out_wg)
	}
	out_wg.Wait()
}

func retryJob(project_id string, job_id int) (job model.Job) {
	path := fmt.Sprintf("projects/%s/jobs/%d/retry", project_id, job_id)
	body := post_no_data(path)

	json.Unmarshal(body, &job)
	fmt.Printf("new retry job, id: %d\n", job.ID)
	return job
}

func cancelJob(project_id string, job_id int) (job model.Job) {
	fmt.Printf("cancelling job: %d\n", job_id)
	path := fmt.Sprintf("projects/%s/jobs/%d/cancel", project_id, job_id)
	body := post_no_data(path)

	json.Unmarshal(body, &job)
	return job
}

func getJob(project_id string, job_id int) (job model.Job) {
	path := fmt.Sprintf("projects/%s/jobs/%d", project_id, job_id)
	body := get(path)

	json.Unmarshal(body, &job)
	return job
}

func getJobsPageCount(id string, count int) (jobs model.Jobs) {
	for i := 0; i < count; i++ {
		new_jobs := getJobsPage(id, i)
		jobs = jobs.Combine(new_jobs)
	}
	return jobs
}

func getJobsPage(id string, page int) (jobs model.Jobs) {
	path := fmt.Sprintf("projects/%s/jobs?page=%d", id, page)
	body := get(path)

	json.Unmarshal(body, &jobs)
	return jobs
}

func getJobs(id string) (jobs model.Jobs) {
	path := fmt.Sprintf("projects/%s/jobs", id)
	body := get(path)

	json.Unmarshal(body, &jobs)
	return jobs
}

func searchProjects(name string) (projects []model.Project) {
	path := fmt.Sprintf("projects?search=%s", name)
	body := get(path)

	json.Unmarshal(body, &projects)
	return projects
}

func post_no_data(path string) (body []byte) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v4/%s", base_url, path)
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Add("PRIVATE-TOKEN", token)
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

func get(path string) (body []byte) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v4/%s", base_url, path)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", token)
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
