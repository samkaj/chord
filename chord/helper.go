package chord

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"math/big"
	"os"
)

const keySize = sha1.Size * 8

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

// Hashes a string to a big.Int using SHA1
func Hash(addr string) *big.Int {
	h := sha1.New()
	h.Write([]byte(addr))
	return new(big.Int).SetBytes(h.Sum(nil))
}

// If inclusive: (start, end]
// Else: (start, end)
func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

// Read a file and return its contents as a byte array
func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %s\n", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanBytes)
	var data []byte
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %s\n", err)
		return nil, err
	}
	return data, nil
}
