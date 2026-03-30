package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) ServeSwagger(c *gin.Context) {
	h.ETL.ServeSwagger(c)
}

