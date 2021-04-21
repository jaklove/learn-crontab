package app

import "github.com/gin-gonic/gin"

// BindAndValid binds
func BindAndValid(c *gin.Context, form interface{}) error {
	err := c.Bind(form)
	if err != nil {
		return err
	}
	return nil
}
