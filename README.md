## Test commands

**Create Ring**
```bash
./chord.sh -a 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

**Join Ring 1**
```bash
./chord.sh -a 0.0.0.0:3001 -j 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

**Join ring 2**
```bash
./chord.sh -a 0.0.0.0:3002 -j 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

**Join ring 3**
```bash
./chord.sh -a 0.0.0.0:3003 -j 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
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
    -addext "subjectAltName = IP:192.168.1.106" \
    -keyout key.pem -out cert.pem

```