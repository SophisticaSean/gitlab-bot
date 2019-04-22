package main

import (
	"fmt"
	"time"

	"github.com/SophisticaSean/gitlab-bot/gitlab"
	"github.com/SophisticaSean/gitlab-bot/model"
)

var app_id = "21"
var job_count = 500

func main() {

	gl := gitlab.New()

	var jobs model.Jobs

	// wait for a new Unit Tests job
	for {
		fmt.Println("waiting for a running/pending new Unit Tests job")
		jobs = gl.GetJobsPageCount(app_id, 5)
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
	gl.CancelJobs(app_id, jobs)

	// wait for a new canceled Unit Tests job
	for {
		fmt.Println("waiting for a new canceled Unit Tests job")
		jobs = gl.GetJobsPageCount(app_id, 5)
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

	jobs = gl.ChunkRetryJobID(app_id, job.ID, job_count)
	if len(jobs) != job_count {
		panic(fmt.Sprintf("jobs is not %d!", job_count))
	}

	// wait for the jobs to complete
	for {
		current_jobs := model.Jobs{}
		for _, job := range jobs {
			job = gl.GetJob(app_id, job.ID)
			current_jobs = append(current_jobs, job)
			// cancel any jobs running longer than 10 minutes
			if job.Duration >= 600.0 {
				gl.CancelJobs(app_id, model.Jobs{job})
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
		fmt.Printf("%s/backend/juno/-/jobs/%d\n", gl.BaseURL, job.ID)
	}
	fmt.Println("")
}
