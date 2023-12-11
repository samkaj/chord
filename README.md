<div align="center">
<h1>Laboration 3</h1>
<h3>Distributed storage using Chord</h3>

[Canvas instructions](https://chalmers.instructure.com/courses/26458/pages/lab-3-chord) | [Chord paper](https://people.eecs.berkeley.edu/~istoica/papers/2003/chord-ton.pdf) | [Tutorial](https://computing.utahtech.edu/cs/3410/asst_chord.html)

Samuel Kajava, Daniel Rygaard, Dennis Sehic
</div>

## General

We use TLS for securely transferring files, and you therefore need to generate a key and a certificate to be able to run the application, see [Creating SSL certificate](#creating-ssl-certificate).

## Test commands

**Build**

```bash
go build -o build/chord
```

**Create Ring**

```bash
build/chord -a 0.0.0.0 -p 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 8081
```

**Join Ring 1**

```bash
build/chord -a 0.0.0.0 -p 2020 -ja 0.0.0.0 -jp 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 2021
```

**Join ring 2**

```bash
build/chord -a 0.0.0.0 -p 3030 -ja 0.0.0.0 -jp 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 3031
```

**Join ring 3**

```bash
build/chord -a 0.0.0.0 -p 4040 -ja 0.0.0.0 -jp 8080 -tcp 1000 -ff 1000 -ts 100 -r 3 -tls 4041
```

## Creating SSL certificate

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
    -subj "/CN=chord-server" \
    -addext "subjectAltName = IP:0.0.0.0" \
    -keyout key.pem -out cert.pem

```
