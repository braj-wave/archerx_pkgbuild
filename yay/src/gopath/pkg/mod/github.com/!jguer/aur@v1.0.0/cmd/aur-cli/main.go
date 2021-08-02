package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Jguer/aur"
)

const (
	boldCode  = "\x1b[1m"
	resetCode = "\x1b[0m"
)

// UseColor determines if package will emit colors.
var UseColor = true // nolint

func getSearchBy(value string) aur.By {
	switch value {
	case "name":
		return aur.Name
	case "maintainer":
		return aur.Maintainer
	case "depends":
		return aur.Depends
	case "makedepends":
		return aur.MakeDepends
	case "optdepends":
		return aur.OptDepends
	case "checkdepends":
		return aur.CheckDepends
	default:
		return aur.NameDesc
	}
}

func usage() {
	fmt.Println("Usage:", os.Args[0], "<opts>", "<command>", "<pkg(s)>")
	fmt.Println("Available commands:", "info, search")
	fmt.Println("Available opts:", "-by <Search for packages using a specified field>")

	flag.Usage()

	fmt.Println("Example:", "aur-cli -verbose -by name search python3.7")
}

func main() {
	var (
		by          string
		aurURL      string
		verbose     bool
		jsonDisplay bool
	)

	flag.StringVar(&by, "by", "name-desc", "Search for packages using a specified field"+
		"\n (name/name-desc/maintainer/depends/makedepends/optdepends/checkdepends)")
	flag.StringVar(&aurURL, "url", "https://aur.archlinux.org/", "AUR URL")
	flag.BoolVar(&verbose, "verbose", false, "display verbose information")
	flag.BoolVar(&jsonDisplay, "json", false, "display result as JSON")
	flag.Parse()

	if flag.NArg() < 2 {
		usage()

		os.Exit(1)
	}

	aurClient, err := aur.NewClient(aur.WithBaseURL(aurURL),
		aur.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", "aur-cli/v1")

			return nil
		}))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	results, err := fn0(aurClient, by)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}

	if jsonDisplay {
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)

			os.Exit(1)
		}

		fmt.Println(string(output))
	} else {
		for i := range results {
			printInfo(&results[i], aurURL, verbose, flag.Arg(0))
		}
	}
}

func fn0(aurClient *aur.Client, by string) (results []aur.Pkg, err error) {
	switch flag.Arg(0) {
	case "search":
		results, err = aurClient.Search(context.Background(), strings.Join(flag.Args()[1:], " "), getSearchBy(by))
	case "info":
		results, err = aurClient.Info(context.Background(), flag.Args()[1:])
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		err = fmt.Errorf("rpc request failed: %w", err)
	}

	return results, err
}

func stylize(startCode, in string) string {
	if UseColor {
		return startCode + in + resetCode
	}

	return in
}

func Bold(in string) string {
	return stylize(boldCode, in)
}

// PrintInfo prints package info like pacman -Si.
func printInfo(a *aur.Pkg, aurURL string, verbose bool, mode string) {
	printInfoValue("Name", a.Name)
	printInfoValue("Version", a.Version)
	printInfoValue("Description", a.Description)

	if verbose {
		printInfoValue("Keywords", a.Keywords...)
		printInfoValue("URL", a.URL)
		printInfoValue("AUR URL", strings.TrimRight(aurURL, "/")+"/packages/"+a.Name)

		if mode == "info" {
			printInfoValue("Groups", a.Groups...)
			printInfoValue("Licenses", a.License...)
			printInfoValue("Provides", a.Provides...)
			printInfoValue("Depends On", a.Depends...)
			printInfoValue("Make Deps", a.MakeDepends...)
			printInfoValue("Check Deps", a.CheckDepends...)
			printInfoValue("Optional Deps", a.OptDepends...)
			printInfoValue("Conflicts With", a.Conflicts...)
		}

		printInfoValue("Maintainer", a.Maintainer)
		printInfoValue("Votes", fmt.Sprintf("%d", a.NumVotes))
		printInfoValue("Popularity", fmt.Sprintf("%f", a.Popularity))
		printInfoValue("First Submitted", formatTimeQuery(a.FirstSubmitted))
		printInfoValue("Last Modified", formatTimeQuery(a.LastModified))

		if a.OutOfDate != 0 {
			printInfoValue("Out-of-date", formatTimeQuery(a.OutOfDate))
		} else {
			printInfoValue("Out-of-date", "No")
		}

		printInfoValue("ID", fmt.Sprintf("%d", a.ID))
		printInfoValue("Package Base ID", fmt.Sprintf("%d", a.PackageBaseID))
		printInfoValue("Package Base", a.PackageBase)
		printInfoValue("Snapshot URL", aurURL+a.URLPath)
	}

	fmt.Println()
}
