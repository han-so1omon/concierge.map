package main

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/securecookie"
	"os"
	"reflect"
)

type keys struct {
	AuthKey       []byte
	EncryptionKey []byte
}

func main() {
	fmt.Println("Generating auth and encryption keys for gorilla session management")

	k := keys{
		AuthKey:       securecookie.GenerateRandomKey(64),
		EncryptionKey: securecookie.GenerateRandomKey(32),
	}

	fname := "tmp.env"
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0755)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	enc := gob.NewEncoder(f)
	if err = enc.Encode(k); err != nil {
		panic(err)
	}

	var kTest keys
	fTest, errTest := os.OpenFile(fname, os.O_RDONLY, 0664)
	defer fTest.Close()
	if errTest != nil {
		panic(errTest)
	}
	dec := gob.NewDecoder(fTest)
	if errTest = dec.Decode(&kTest); errTest != nil {
		panic(errTest)
	}

	if !reflect.DeepEqual(k, kTest) {
		fmt.Println("Keys written do not match keys read")
	}

	fmt.Println("Keys written to ", fname)
	fmt.Println("Rename keys file as necessary. suggested name is .session_keys")
}
