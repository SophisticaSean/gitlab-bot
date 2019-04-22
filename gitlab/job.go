package gitlab

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/SophisticaSean/gitlab-bot/model"
)

func (gitlab Gitlab) ChunkRetryJobID(project_id string, job_id int, count int) model.Jobs {
	jobs := model.Jobs{}
	chunk_by := 25
	chunk_count := count / chunk_by
	remainder := count % chunk_by
	// retry jobs in chunks of 50
	for i := 0; i < chunk_count; i++ {
		new_jobs := gitlab.retryJobID(project_id, job_id, chunk_by)
		jobs = jobs.Combine(new_jobs)
		time.Sleep(3 * time.Second)
	}
	// handle remainder
	if remainder > 0 {
		new_jobs := gitlab.retryJobID(project_id, job_id, remainder)
		jobs = jobs.Combine(new_jobs)
	}
	return jobs
}

// retry job_id in project_id count times
//  return the array of new jobs
func (gitlab Gitlab) retryJobID(project_id string, job_id int, count int) model.Jobs {
	out_wg := sync.WaitGroup{}
	job_slice := model.JobSlice{}
	for i := 0; i < count; i++ {
		out_wg.Add(1)
		go gitlab.retryJobAsync(project_id, job_id, &out_wg, &job_slice)
	}
	out_wg.Wait()
	return job_slice.Slice
}

func (gitlab Gitlab) retryJobAsync(project_id string, job_id int, new_wg *sync.WaitGroup, js *model.JobSlice) {
	defer new_wg.Done()
	defer js.Unlock()
	new_job := gitlab.RetryJob(project_id, job_id)

	// lock the mutex and append our new job
	js.Lock()
	js.Slice = append(js.Slice, new_job)
}

func (gitlab Gitlab) CancelJobs(project_id string, jobs model.Jobs) {
	out_wg := sync.WaitGroup{}
	for _, j := range jobs {
		out_wg.Add(1)
		go func(job model.Job, wg *sync.WaitGroup) {
			defer wg.Done()
			gitlab.cancelJob(project_id, job.ID)
		}(j, &out_wg)
	}
	out_wg.Wait()
}

func (gitlab Gitlab) RetryJob(project_id string, job_id int) (job model.Job) {
	path := fmt.Sprintf("projects/%s/jobs/%d/retry", project_id, job_id)
	body := gitlab.PostNoData(path)

	json.Unmarshal(body, &job)
	fmt.Printf("new retry job, id: %d\n", job.ID)
	return job
}

func (gitlab Gitlab) cancelJob(project_id string, job_id int) (job model.Job) {
	fmt.Printf("cancelling job: %d\n", job_id)
	path := fmt.Sprintf("projects/%s/jobs/%d/cancel", project_id, job_id)
	body := gitlab.PostNoData(path)

	json.Unmarshal(body, &job)
	return job
}

func (gitlab Gitlab) GetJob(project_id string, job_id int) (job model.Job) {
	path := fmt.Sprintf("projects/%s/jobs/%d", project_id, job_id)
	body := gitlab.Get(path)

	json.Unmarshal(body, &job)
	return job
}

func (gitlab Gitlab) GetJobsPageCount(id string, count int) (jobs model.Jobs) {
	for i := 0; i < count; i++ {
		new_jobs := gitlab.getJobsPage(id, i)
		jobs = jobs.Combine(new_jobs)
	}
	return jobs
}

func (gitlab Gitlab) getJobsPage(id string, page int) (jobs model.Jobs) {
	path := fmt.Sprintf("projects/%s/jobs?page=%d", id, page)
	body := gitlab.Get(path)

	json.Unmarshal(body, &jobs)
	return jobs
}

func (gitlab Gitlab) getJobs(id string) (jobs model.Jobs) {
	path := fmt.Sprintf("projects/%s/jobs", id)
	body := gitlab.Get(path)

	json.Unmarshal(body, &jobs)
	return jobs
}
