package controllers // it has api request handlers

import (
	"encoding/json"
	// "fmt"
	"net/http"
	"olake_backend/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type SourceController struct {
	web.Controller 
	o orm.Ormer
}

func (c *SourceController) Prepare() {
	c.o = orm.NewOrm() 
}

func (c *SourceController) CreateSource() {
	var source models.Source
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &source); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	currentTime := time.Now()
	source.CreatedAt = currentTime 
	source.UpdatedAt = currentTime 

	_, err := c.o.Insert(&source)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to create source"}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = source
	c.ServeJSON()
}

// GetAllSources retrieves all sources
func (c *SourceController) GetAllSources() {
	var sources []models.Source
	_, err := c.o.QueryTable("source").All(&sources)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError) 
		c.Data["json"] = map[string]string{"error": "Failed to retrieve sources"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = sources
	c.ServeJSON()
}

// UpdateSource updates an existing source
func (c *SourceController) UpdateSource() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}

	var source models.Source

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &source); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	_, err = c.o.QueryTable("source").Filter("source_id", id).Update(orm.Params{
		"name":        source.Name,
		"project_id":  source.ProjectID,
		"config":      source.Config,
		"updated_at":  time.Now(),
		"updated_by":  source.UpdatedBy,
		"source_type": source.SourceType,
	})
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to update source"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = source
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.ServeJSON()
}

// DeleteSource deletes a source
func (c *SourceController) DeleteSource() {
	
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}
	_, err = c.o.Delete(&models.Source{SourceID: int(id)})
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to delete source"}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusNoContent)
} 