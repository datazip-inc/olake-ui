package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"olake_backend/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type JobController struct {
	web.Controller
	o orm.Ormer
}

// Prepare initializes the ORM instance
func (c *JobController) Prepare() {
	c.o = orm.NewOrm()
}

// CreateJob handles the creation of a new job
func (c *JobController) CreateJob() {
	var jobRequest struct {
		SourceID      int    `json:"source_id"`
		DestID        int    `json:"dest_id"`
		Name          string `json:"name"`
		Connected     bool   `json:"connected"`
		Frequency     string `json:"frequency"`
		Config        string `json:"config"`
		CreatedBy     string `json:"created_by"`
		UpdatedBy     string `json:"updated_by"`
		LastSyncState map[string]interface{} `json:"last_sync_state"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &jobRequest); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	// Create new job
	job := &models.Job{
		SourceID:     jobRequest.SourceID,
		DestID:       jobRequest.DestID,
		Name:         jobRequest.Name,
		Connected:    jobRequest.Connected,
		Frequency:    jobRequest.Frequency,
		Config:       jobRequest.Config,
		CreatedBy:    jobRequest.CreatedBy,
		UpdatedBy:    jobRequest.UpdatedBy,
	}

	// Set timestamps
	currentTime := time.Now()
	job.CreatedAt = currentTime
	job.UpdatedAt = currentTime

	// Set LastSyncState
	if lastRunTime, ok := jobRequest.LastSyncState["last_run_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastRunTime); err == nil {
			job.LastRunTime = &t
		}
	}
	if lastRunState, ok := jobRequest.LastSyncState["last_run_state"].(string); ok {
		job.LastRunState = lastRunState
	}

	// Save job to the database
	_, err := c.o.Insert(job)
	if err != nil {
		fmt.Println("Error creating job:", err)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to create job"}
		c.ServeJSON()
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"job_id":      job.JobID,
		"source_id":   job.SourceID,
		"dest_id":     job.DestID,
		"name":        job.Name,
		"connected":   job.Connected,
		"created_at":  job.CreatedAt,
		"updated_at":  job.UpdatedAt,
		"last_sync_state": map[string]interface{}{
			"last_run_time":  job.LastRunTime,
			"last_run_state": job.LastRunState,
		},
		"frequency":   job.Frequency,
		"config":      job.Config,
		"created_by":  job.CreatedBy,
		"updated_by":  job.UpdatedBy,
		"deleted_at":  job.DeletedAt,
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = response
	c.ServeJSON()
}

// GetAllJobs retrieves all jobs
func (c *JobController) GetAllJobs() {
	var jobs []models.Job
	
	// Initialize response slice with empty array
	response := make([]map[string]interface{}, 0)
	
	_, err := c.o.QueryTable("job").All(&jobs)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to fetch jobs"}
		c.ServeJSON()
		return
	}

	// Convert jobs to response format if there are any jobs
	for _, job := range jobs {
		jobMap := map[string]interface{}{
			"job_id":      job.JobID,
			"source_id":   job.SourceID,
			"dest_id":     job.DestID,
			"name":        job.Name,
			"connected":   job.Connected,
			"created_at":  job.CreatedAt,
			"updated_at":  job.UpdatedAt,
			"last_sync_state": job.GetLastSyncState(),
			"frequency":   job.Frequency,
			"config":      job.Config,
			"created_by":  job.CreatedBy,
			"updated_by":  job.UpdatedBy,
			"deleted_at":  job.DeletedAt,
		}
		response = append(response, jobMap)
	}

	// Will return empty array if no jobs found
	c.Data["json"] = response
	c.ServeJSON()
}

// UpdateJob updates an existing job
func (c *JobController) UpdateJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}

	var jobRequest struct {
		SourceID      int    `json:"source_id"`
		DestID        int    `json:"dest_id"`
		Name          string `json:"name"`
		Connected     bool   `json:"connected"`
		Frequency     string `json:"frequency"`
		Config        string `json:"config"`
		UpdatedBy     string `json:"updated_by"`
		LastSyncState map[string]interface{} `json:"last_sync_state"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &jobRequest); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	// Parse LastSyncState
	var lastRunTime *time.Time
	if lastRunTimeStr, ok := jobRequest.LastSyncState["last_run_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastRunTimeStr); err == nil {
			lastRunTime = &t
		}
	}

	// Update the job
	_, err = c.o.QueryTable("job").Filter("job_id", id).Update(orm.Params{
		"source_id":      jobRequest.SourceID,
		"dest_id":        jobRequest.DestID,
		"name":          jobRequest.Name,
		"connected":     jobRequest.Connected,
		"frequency":     jobRequest.Frequency,
		"config":        jobRequest.Config,
		"updated_by":    jobRequest.UpdatedBy,
		"updated_at":    time.Now(),
		"last_run_time": lastRunTime,
		"last_run_state": jobRequest.LastSyncState["last_run_state"],
	})
if err != nil {
		fmt.Println("Error updating job:", err)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to update job"}
		c.ServeJSON()
		return
	}

	// Fetch the updated job
	updatedJob := models.Job{JobID: id}
	err = c.o.Read(&updatedJob)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to fetch updated job"}
		c.ServeJSON()
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"job_id":      updatedJob.JobID,
		"source_id":   updatedJob.SourceID,
		"dest_id":     updatedJob.DestID,
		"name":        updatedJob.Name,
		"connected":   updatedJob.Connected,
		"created_at":  updatedJob.CreatedAt,
		"updated_at":  updatedJob.UpdatedAt,
		"last_sync_state": map[string]interface{}{
			"last_run_time":  updatedJob.LastRunTime,
			"last_run_state": updatedJob.LastRunState,
		},
		"frequency":   updatedJob.Frequency,
		"config":      updatedJob.Config,
		"created_by":  updatedJob.CreatedBy,
		"updated_by":  updatedJob.UpdatedBy,
		"deleted_at":  updatedJob.DeletedAt,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// DeleteJob deletes a job
func (c *JobController) DeleteJob() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}

	// Delete the job
	_, err = c.o.Delete(&models.Job{JobID: id})
	if err != nil {
		fmt.Println("Error deleting job:", err)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to delete job"}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusNoContent)
} 