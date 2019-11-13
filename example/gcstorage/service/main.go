package main

import (
	"context"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
)

func main() {
	time.Sleep(time.Second * 2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	buff, err := read(ctx, "bucket/path", "file.txt")
	if err != nil {
		panic(err)
	}

	if err := write(ctx, "bucket/path/bak", "file.txt.bak", buff); err != nil {
		panic(err)
	}
}

func read(ctx context.Context, bucket, file string) ([]byte, error) {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	object := c.Bucket("bucket/path").Object("file.txt")
	r, err := object.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return buff, err
}

func write(ctx context.Context, bucket, file string, buff []byte) error {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	object := c.Bucket(bucket).Object(file)
	w := object.NewWriter(ctx)
	if _, err = w.Write(buff); err != nil {
		return err
	}
	return w.Close()
}