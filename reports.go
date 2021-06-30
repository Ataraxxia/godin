package main

import (
	"net/url"
	"strings"
	"encoding/json"
	"time"
	"fmt"
	"os"
	"strconv"
	"io/ioutil"
        log "github.com/sirupsen/logrus"
)

// TODO Optimize args and returns, switch to slices and pointers

func parsePackages(packages []string) []Package {
	var pkgs []Package

        for _, pack := range packages {
                pack = strings.ReplaceAll(pack, "'", "")
                p := strings.Split(pack, " ")
                if len(p) <= 1 {
                        continue
                }

                var pkg Package
                pkg.Name = p[0]
                pkg.Epoch = p[1]
                pkg.Version = p[2]
                pkg.Release = p[3]
                pkg.Arch = p[4]
                pkg.PkgManager = p[5]

                pkgs = append(pkgs, pkg)
        }
	return pkgs
}

func parseRepositories(repositories []string) []Repository {
	var repos []Repository

	for _, repo := range repositories {
		r := strings.Split(repo, "' '")
		for i, item := range r {
			item = strings.ReplaceAll(item, "'", "")
			r[i] = item
		}
		if len(r) == 0 {
			continue
		}

		var repo Repository
		repo.Type = r[0]

		switch repoType := repo.Type; repoType {
		case "deb":
			repo.Name = r[1]
			repo.Priority, _ = strconv.Atoi(r[2])
			repo.Url = strings.ReplaceAll(r[3], " " , "")
		case "rpm":
			repo.Name = r[1]
			repo.Priority, _ = strconv.Atoi(r[3])
			repo.Url = strings.ReplaceAll(r[len(r) - 1], " ", "") //The last element in patchman report tends to be actuall repository url

		default:
			log.Debugf("Unknown repository type %s", repoType)
			continue
		}

		repos = append(repos, repo)
	}
	return repos
}

func saveReport(form url.Values) error {

	var report Report
        report.Host = form["host"][0]
        report.Tags = strings.Split(form["tags"][0], " ")
        report.Kernel = form["kernel"][0]
        report.Arch = form["arch"][0]
        report.Protocol = form["protocol"][0]
        report.OS = form["os"][0]
        report.Reboot = form["reboot"][0]

        packages := strings.Split(form["packages"][0], "\n")
	report.Packages = parsePackages(packages)

	secUpdates := strings.Split(form["sec_updates"][0], "\n")
        report.SecUpdates = parsePackages(secUpdates)

	bugUpdates := strings.Split(form["bug_updates"][0], "\n")
        report.BugUpdates = parsePackages(bugUpdates)

	repos := strings.Split(form["repos"][0], "\n")
        report.Repos = parseRepositories(repos)

        bytes, err := json.Marshal(report)
        if err != nil {
                log.Error(err)
        }

        dt := strings.ReplaceAll(time.Now().Format("01-02-2006 15:04:05"), " ", "_")

        fpath := fmt.Sprintf("%s/%s_%s.json", config.DataPath, report.Host, dt)
        ioutil.WriteFile(fpath, bytes, os.ModePerm)

	return nil
}
