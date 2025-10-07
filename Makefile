update-postman-collection: backend/docs/swagger.yaml 
	@openapi2postmanv2 -s backend/docs/swagger.yaml -o postman/kubecloud_collection.json -p