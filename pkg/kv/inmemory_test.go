package kv_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/shivamhw/content-pirate/pkg/kv"
)

var db *kv.InMemDb

func setup(){
	db = kv.GetInMemoryKv()
}


func TestDel(t *testing.T){
	setup()
	err := db.Set("task", "123", []byte("details"))
	if err != nil {
		t.Fatalf("failed in get %s", err)
	}
	val, err := db.Get("task", "123")
	if err != nil {
		t.Fatal("failed in get")
	}

	fmt.Printf("%s", string(val))
}

func TestGet(t *testing.T){
	setup()
	err := db.Set("task", "123", []byte("details"))
	if err != nil {
		log.Fatalf("failed in get %s", err)
	}
	val, err := db.Get("task", "123")
	if err != nil {
		log.Fatal("failed in get")
	}
	fmt.Printf("%s", string(val))
	err = db.Del("task", "123")
	if err != nil {
		t.Fatalf("failed in del %s", err)
	}
	val, err = db.Get("task", "123")
	if err != nil {
		t.Logf("passed as err is %s", err)
		fmt.Print(err)
	} else {
		t.Fatalf("deleted value found %s", string(val))
	}
}
