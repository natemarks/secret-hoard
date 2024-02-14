# secret-hoard


This project is used to create or update AWS secretsmanager secrets. A secret is used to grant a specific 'access' to a specific resource. As an example, an RDS secret might grant read-only access to a specific database in a specific environment.

### secret upload executables

There are five different types of secrets that can be created or updated using this project. Each type of secret has a different format and different metadata. Each type hase a unique executable to upload secrets from a csv input file. 

All upload executables iterate through a CSV file and create secrets that don't exist, but leave untouched secrets that do exist. The overwrite flag can be used to update existing secrets.

#### upload rdspostgres secrets
rdspostgres secrets grant access to an RDS instance and database.

```bash
sh-rdspostgres -file=examples/rdspostgres_example.csv -debug -overwrite
```

This is an example CSV file with a single entry
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
  "Access": "mytype",
  "Source": "secret-hoard"
}
```

The secret value will contain the information required to access the resource. For a rdspostgres, it would contain the host, port, username, password, etc.  in a predictable format so the secret can be used to build a connection string, rotated, etc.

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

#### upload snowflake secrets
snowflake secrets grant access to a snowflake.

```bash
sh-snowflake -file=examples/snowflake_example.csv -debug -overwrite
```

The CSV will look like:
```csv
ResourceType,Environment,Warehouse,Access,AccountName,Username,Password
snowflake,myenvironment,mywarehouse,mytype,myAccountname,myusername,mypassword
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
  "Access": "read",
  "Source": "secret-hoard"
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


#### upload ssl certificate secrets

ssl_certificate secrets store the certificate and private key file for a common name. The csv file contains paths to the two files and the secret includes information about the files like sha256sum, expiration date and modulus.

```bash
sh-sslcert -file=examples/sslcert_example.csv -debug -overwrite
```

The CSV will look like:

```csv
ResourceType,Environment,CommonName,CertificateFile,PrivateKeyFile
ssl_certificate,testenv,my.domain.com,examples/certificate.crt,examples/private_key.key
```


The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[CommonName]
Secret ID Example:  'ssl_certificate/testenv/my.domain.com'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "ssl_certificate",
  "Environment": "testenv",
  "CommonName": "my.domain.com",
  "Source": "secret-hoard"
}
```

The secret value will contain the information required to access the resource. For an RDS instance, it would contain the
host, port, username, password, etc.  in a predictable format so the secret can be used, rotated, etc.

```json
{
  "certificate": "...",
  "key": "...",
  "expirationDate": "...",
  "modulus": "...",
  "certificateSha256": "...",
  "privateKeySha256": "..."
}
```





#### upload jsondoc secrets
jsondoc secrets store a json document and the sha2456sum of the document.

```bash
sh-jsondoc -file=examples/jsondoc_example.csv -debug -overwrite
```

The CSV will look like:
```csv
ResourceType,Environment,Access,File
jsondoc,testenv,some_json_access_type,examples/jsondoc_example.json
```

The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[Access]
Secret ID Example:  'jsondoc/testenv/some_json_access_type'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "jsondoc",
  "Environment": "testenv",
  "CommonName": "some_json_access_type",
  "Source": "secret-hoard"
}
```

The secret value will contain the information required recreate the file. The contents are stored as a string and the sha256sum can be used to validate the download.

```json
{
  "JSONContents": "...",
  "JSONSha256Sum": "..."
}
```



## upload textfile secrets
textfile secrets store a text file and the sha2456sum of the file.

```bash
sh-textfile -file=examples/textfile_example.csv -debug -overwrite
```

The CSV will look like:
```csv
ResourceType,Environment,Access,FilePath
text_file,testenv,my_file_type,examples/text_file_example.txt
```


The secret ID will be formed from the metadata
```text
Secret ID Format: [ResourceType]/[Environment]/[Access]
Secret ID Example:  'textfile/testenv/my_file_type'
```

The Tags will be set based on the metadata
```json
{
  "ResourceType": "textfile",
  "Environment": "testenv",
  "CommonName": "my_file_type",
  "Source": "secret-hoard"
}
```

The secret value will contain the information required recreate the file. The contents are stored as a string and the sha256sum can be used to validate the download.


```json
{
  "JSONContents": "...",
  "JSONSha256Sum": "..."
}
```


## download secrets
The 'sh-get' executable can be used to download any one secret by its secret ID to a given file path. For most secrets sh-get downloads the contents to a single file.
```bash
sh-get -id=rdspostgres/testenv/myinstance/mydb/mytype -file=private/rdspostgres_testenv_myinstance_mydb_mytype.json -debug
```

For sslcert secrets, sh-get downloads two files using the given filepath string as a prefix. The files are named with .key and .crt extensions. 

In this example: my_domain.crt and my_domain.key
```bash
sh-get -id=sslcert/testenv/my.domain.com-file=private/my_domain -debug
```

