# secret-hoard
This project is used to create AWS secretsmanager secrets. A secret is used to grant a specific 'access' to a specific resource. As an example, an RDS secret might grant read-only access to a specific database in a specific environment.  The secret value will contain the data required to connect to the database:

host, port, username, password, etc.

The metadata would contain information about the asset and access. The metadata is important to automate management, rotation, etc.
resource type, environment, instance, database, access

The metadata will describe the assets and access granted by the secret so the can be managed, rotated, etc.



The metadata for a secret should include the environment, assets to which the secrets grants access and an access profile that describes the type of access granted. It DOES NOT include versioning information required to rotate the secret. Rotation will use the internal secret versioning mechanism.

## metadata examples

### Read-only access ot a Postgres RDS database
 A secret that grants readonly access to a database named 'tires' on the RDS instance 'factory' in the dev environment would have the following metadata
 ```json
 {
    "resourceType": "rdspostgres",
    "environment": "dev",
    "instance": "factory",
    "database": "tires",
    "access": "readonly"
 }
```
We'll limit the metadata value characters to characters  permitted in secret IDs so the names can be  a concatenation of the metadata values.  The secret id would be rdspostgres/dev/factory/tires/readonly. NOTE: the secret id can be 512 characters

### Master access to a Postgres RDS instance

Metadata for a secret containing teh postgres master instance credentials for teh factory instance in the dev environemtn
```

 ```json
{
    "resourceType": "rdspostgres",
    "environment": "dev",
    "instance": "factory",
    "database": "",
    "access": "master"
 }
```
The secret id would be rdspostgres/dev/factory/master

### readwrite access to a snowflake warehouse
```json
{
    "resourceType": "snowflake",
    "environment": "dev",
    "warehouse": "my_warehouse",
    "access": "readwrite"
 }
```


### SSL certificate file

We usually  need to store a certificate file and a certificate key file:
```json
{
    "resourceType": "sslcert",
    "environment": "dev",
    "domain": "my_domain",
    "access": "certificate"
 }
```
```json
{
    "resourceType": "sslcert",
    "environment": "dev",
    "domain": "my_domain",
    "access": "key"
 }
```


### arbitrary file

Sometimes we need to store some arbitrary file.

```json
{
  "resourceType": "file",
  "environment": "dev",
  "domain": "file_name"
}
```



## secret value


The secret value must permit rotation, so it must include all the information required to connect to the resource and rotate credentials.  For an RDS Postgres database, for example, it would include the username, password, host, port, etc.  For snowflake, it would contain username, password, account and warehouse. 


