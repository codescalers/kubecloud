package main

import "kubecloud/cmd"

// @title Mycelium Cloud API
// @version 1.0
// @description API documentation for Mycelium Cloud.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cmd.Execute()
}
