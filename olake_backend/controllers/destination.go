package controllers

import (
	"encoding/json"
	"net/http"
	"olake_backend/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DestinationController struct {
	web.Controller
	o orm.Ormer
}

func (c *DestinationController) Prepare() {
	c.o = orm.NewOrm() 
}

// CreateDestination handles the creation of a new destination
func (c *DestinationController) CreateDestination() {
	var destination models.Destination
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &destination); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	currentTime := time.Now()
	destination.CreatedAt = currentTime 
	destination.UpdatedAt = currentTime 
	
	_, err := c.o.Insert(&destination)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to create destination"}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = destination
	c.ServeJSON()
}

// GetAllDestinations retrieves all destinations
func (c *DestinationController) GetAllDestinations() {
	var destinations []models.Destination
	_, err := c.o.QueryTable("destination").All(&destinations)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to retrieve destinations"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = destinations
	c.ServeJSON()
}

// UpdateDestination updates an existing destination
func (c *DestinationController) UpdateDestination() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}
	var destination models.Destination
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &destination); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	// Use id to find the destination in the database
	_, err = c.o.QueryTable("destination").Filter("dest_id", id).Update(orm.Params{
		"name":        destination.Name,
		"project_id":  destination.ProjectID,
		"config":      destination.Config,
		"updated_at":  time.Now(),
		"updated_by":  destination.UpdatedBy,
		"dest_type":   destination.DestType,
	})
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to update destination"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = destination
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.ServeJSON()
}

// DeleteDestination deletes a destination
func (c *DestinationController) DeleteDestination() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}
	_, err = c.o.Delete(&models.Destination{DestID: int(id)}) 
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to delete destination"}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusNoContent)
} 