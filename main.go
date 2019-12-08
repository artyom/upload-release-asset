// TODO describe program
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	args := runArgs{
		User:      os.Getenv("GITHUB_ACTOR"),
		Token:     os.Getenv("GITHUB_TOKEN"),
		UploadURL: os.Getenv("INPUT_UPLOAD_URL"),
	}
	// autoflags.Parse(&args)
	flag.StringVar(&args.User, "user", args.User, "github user")
	flag.StringVar(&args.Token, "token", args.Token, "github authorization token")
	flag.StringVar(&args.UploadURL, "url", args.UploadURL, "release assets upload url")
	assets := flag.Args()
	if s := os.Getenv("INPUT_ASSETS"); s != "" {
		assets = filepath.SplitList(s)
	}
	if err := run(args, assets...); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

type runArgs struct {
	User      string `flag:"user,github user"`
	Token     string `flag:"token,github authorization token"`
	UploadURL string `flag:"url,release assets upload url"`
}

func run(args runArgs, assets ...string) error {
	if len(assets) == 0 {
		return errors.New("nothing to upload")
	}
	if args.User == "" {
		return errors.New("empty username")
	}
	if args.Token == "" {
		return errors.New("empty auth token")
	}
	if args.UploadURL == "" {
		return errors.New("empty upload url")
	}
	// github.com/actions/create-release has its outputs.upload_url as
	// https://uploads.github.com/repos/.../assets{?name,label} — need to
	// remove that suffix to get usable url
	args.UploadURL = strings.TrimSuffix(args.UploadURL, `{?name,label}`)
	for _, file := range assets {
		if err := upload(args, file); err != nil {
			return fmt.Errorf("%q upload: %w", file, err)
		}
	}
	return nil
}

func upload(args runArgs, file string) error {
	u, err := url.Parse(args.UploadURL)
	if err != nil {
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	vals := make(url.Values)
	vals.Set("name", filepath.Base(file))
	u.RawQuery = vals.Encode()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), f)
	if err != nil {
		return err
	}
	if fi, err := f.Stat(); err == nil {
		req.ContentLength = fi.Size()
	}
	req.SetBasicAuth(args.User, args.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response status: %q", resp.Status)
	}
	return nil
}
