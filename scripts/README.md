
to pack up the contents of the private/ directory into an encrypted tarball file in secure/:
```bash
bash scripts/secure_private_data.sh
```

to unpack a gpg file in the secure/ directory into the private/ directory:
```bash
bash scripts/unsecure_private_data.sh 20240207-105044.tar.gz.gpg

```


to upload files:
```bash
aws s3 sync secure/ s3://com.imprivata.468716396736.us-east-1.devops-artifacts/secret-hoard/
```