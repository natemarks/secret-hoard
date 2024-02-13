# secret-hoard


This project is used to create or update AWS secretsmanager secrets. A secret is used to grant a specific 'access' to a specific resource. As an example, an RDS secret might grant read-only access to a specific database in a specific environment.

There are five different types of secrets that can be created or updated using this project. Each type of secret has a different format and different metadata. Each type hase a unique executable to upload secrets from a csv input file. There is also an sh-get executable that can be used to download any one secret by its secret ID to a given file path.

## upload dbinstance secrets
```bash
sh-dbinstance -file=examples/dbinstance_example.csv -debug -overwrite
```

## upload snowflake secrets
```bash
sh-snowflake -file=examples/snowflake_example.csv -debug -overwrite
```

## upload ssl certificate secrets
```bash
sh-sslcert -file=examples/sslcert_example.csv -debug -overwrite
```

## upload jsondoc secrets
```bash
sh-jsondoc -file=examples/jsondoc_example.csv -debug -overwrite
```

## upload textfile secrets
```bash
sh-textfile -file=examples/textfile_example.csv -debug -overwrite
```

## download secrets
For most secrets sh-get downloads the contents to a single file.
```bash
sh-get -id=rdspostgres/testenv/myinstance/mydb/mytype -file=private/rdspostgres_testenv_myinstance_mydb_mytype.json -debug
```

For sslcert secrets, sh-get downloads two files using teh given filepath string as a prefix. the files are named with .key and .crt extensions. 

In this example: my_domain.crt and my_domain.key
```bash
sh-get -id=sslcert/testenv/my.domain.com-file=private/my_domain -debug
```


## rdspostgres Secrets

rdspostgres secrets grant access to an RDS instance and database. The source data is stored in a CSV file.
```csv
ResourceType,Environment,Instance,Database,Access,Password,Engine,Port,DbInstanceIdentifier,Host,Username
rdspostgres,testenv,myinstance,mydb,mytype,password,postgres,5432,dbInstanceIdentifier,host,username
```


The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[Instance]/[Database]/[Access]
Secret ID Example:  'rdspostgres/testenv/myinstance/mydb/mytype'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "rdspostgres",
  "Environment": "testenv",
  "Instance": "myinstance",
  "Database": "mydb",
  "Access": "mytype"
}
```

The secret value will contain the information required to access the resource. For an RDS instance, it would contain the
host, port, username, password, etc.  in a predictable format so the secret can be used, rotated, etc.

```json
{
  "password": "password",
  "engine": "postgres",
  "port": 5432,
  "dbInstanceIdentifier": "dbInstanceIdentifier",
  "host": "host",
  "username": "username"
}
```

## snowflake Secrets

rdspostgres secrets grant access to an RDS instance and database. The source data is stored in a CSV file.
```csv
ResourceType,Environment,Warehouse,Access,Password,AccountName,Username
snowflake,testenv,warehouse,read,mytype,password,accountname,username
```


The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[Warehouse]/[Access]
Secret ID Example:  'snowflake/testenv/warehouse/read'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "snowflake",
  "Environment": "testenv",
  "Warehouse": "warehouse",
  "Access": "read"
}
```

The secret value will contain the information required to access the resource. For an RDS instance, it would contain the
host, port, username, password, etc.  in a predictable format so the secret can be used, rotated, etc.

```json
{
  "password": "password",
  "accountName": "accountname",
  "warehouse": "warehouse",
  "username": "username"
}
```



## keyfile Secrets

keyfile secrets contain an SSL key file  matched to its certificate file using the modulus. The csv contains the metadata and a path to the keyfile to be stored in the secret value.
```csv
ResourceType,Environment,CommonName,Modulus,FilePath
keyfile,testenv,my.domain.com,abcdef123456,private/my_domain.key
```


The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[CommonName]/[Modulus]
Secret ID Example:  'keyfile/testenv/my.domain.com/abcdef123456'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "keyfile",
  "Environment": "testenv",
  "CommonName": "my.domain.com",
  "Modulus": "abcdef123456"
}
```

The secret value will contain the text of the key file.
```text
-----BEGIN RSA PRIVATE KEY-----

...

-----END RSA PRIVATE KEY-----
```





## jsonfile Secrets

jsonfile secrets contain an arbitrary json file. The csv contains the metadata and a path to the json file to be stored in the secret value.
```csv
ResourceType,Environment,Access,FilePath
jsonfile,testenv,my_endpoints,private/testenv_my_endpoints.json
```


The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[Access]
Secret ID Example:  'jsonfile/testenv/my_endpoints'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "jsonfile",
  "Environment": "testenv",
  "Access": "my_endpoints"
}
```

The secret value will contain arbitrary, valid JSON






##  usage: rds-hoard
reads a CSV file and creates or updates secrets in AWS Secrets Manager

use the optional -debug flag for troubleshooting
use the optional overwrite flag to overwrite existing secrets

```bash
rds-hoard -file=private/rds_secrets.csv -debug
```