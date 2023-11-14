csv to JSON to csv
csv to secrets to csv


for versio 1 and versio 2  secret structts WITH metadata
instnae

Metadata:
env_id: , rds_instance, release_id, database, credential_type
resource_type:
release_id
database
credential_type








Version 1 secret
```go
// RDSSecret is the struct of the secret generated for RDS by CDK deployment
type RDSSecret struct {
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Port                 int    `json:"port"`
	DbInstanceIdentifier string `json:"dbInstanceIdentifier"`
	Host                 string `json:"host"`
	Username             string `json:"username"`
}


```