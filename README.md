# Test commands

**Build**
```bash
go build -o build/chord
```


**Create Ring**
```bash
build/chord -a localhost -p 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 127.0.0.1:8081
```

**Join Ring 1**
```bash
build/chord -a localhost -p 2020 -ja localhost -jp 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 127.0.0.1:2021
```

**Join ring 2**
```bash
build/chord -a localhost -p 3030 -ja localhost -jp 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 127.0.0.1:3031
```

**Join ring 3**
```bash
build/chord -a localhost -p 4040 -ja localhost -jp 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 127.0.0.1:4041
```

# Creating SSL certificate

Run the following command in the root of the project 
```bash
openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
    -subj "/CN=your-server-common-name" \
    -addext "subjectAltName = IP:<Address without port>" \
    -keyout key.pem -out cert.pem

```
Example:
```bash
openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
    -subj "/CN=your-server-common-name" \
    -addext "subjectAltName = IP:127.0.0.1" \
    -keyout key.pem -out cert.pem

```
