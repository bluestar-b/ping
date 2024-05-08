package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)



func getRecordsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, pingData)
}

func getRecordByIDHandler(c *gin.Context) {
	id := c.Param("id")
	recordID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	if data, ok := pingData[recordID]; ok {
		c.JSON(http.StatusOK, data)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
	}
}
