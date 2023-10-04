package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

func main() {
	var (
		owner string
		repos []string
	)

	pflag.StringVar(&owner, "owner", owner, "Default owner to use for repos not in the OWNER/NAME format.")
	pflag.StringSliceVar(&repos, "repos", repos, "List of repos to query. May specify as NAME or OWNER/NAME. If OWNER is omitted, falls back to --owner.")

	pflag.Parse()

	if len(repos) == 0 {
		log.Println("At least 1 repo must be specified")
		pflag.Usage()
		os.Exit(1)
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	ctx := context.Background()

	stats := make([]*repoStats, 0, len(repos))

	for _, repo := range repos {
		ownerToCheck := owner
		repoToCheck := repo
		if strings.Contains(repo, "/") {
			parts := strings.Split(repo, "/")
			ownerToCheck = parts[0]
			repoToCheck = parts[1]
		} else if ownerToCheck == "" {
			log.Printf("Skipping repo %q because --owner is unset", repo)
			continue
		}

		s, err := getRepoStats(ctx, client, ownerToCheck, repoToCheck)
		if err != nil {
			log.Printf("Error getting stats for %s/%s: %v", ownerToCheck, repoToCheck, err)
			continue
		}
		if s != nil {
			stats = append(stats, s)
		}
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Repo\tx.y.0 releases\tMin days between\tAvg days between\tMax days between\tStdDev\n")
	for _, s := range stats {
		fmt.Fprintf(tw, "%s/%s\t%d\t%.2f\t%.2f\t%.2f\t%.2f\n", s.owner, s.repo, s.dotZeroVersions, s.minDaysBetween, s.averageDaysBetween, s.maxDaysBetween, s.standardDeviation)
	}
	tw.Flush()
}

type repoStats struct {
	owner              string
	repo               string
	dotZeroVersions    int
	minDaysBetween     float64
	averageDaysBetween float64
	maxDaysBetween     float64
	standardDeviation  float64
}

func getRepoStats(ctx context.Context, client *githubv4.Client, owner, repo string) (*repoStats, error) {
	log.Printf("Getting stats for %s/%s\n", owner, repo)
	releases, err := fetchMajorReleases(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	var deltas []time.Duration

	var prev release
	for _, r := range releases {
		if prev.Name == "" {
			prev = r
			continue
		}

		parts := strings.Split(r.Name, ".")
		if len(parts) < 3 || parts[2] != "0" {
			continue
		}

		delta := r.PublishedAt.Sub(prev.PublishedAt)
		deltas = append(deltas, delta)

		prev = r
	}

	var min, max time.Duration
	var sum int64
	for _, d := range deltas {
		if min == 0 {
			min = d
		}
		if d < min {
			min = d
		}

		if max == 0 {
			max = d
		}
		if d > max {
			max = d
		}

		sum += int64(d)
	}

	if len(deltas) == 0 {
		log.Printf("Unable to find any x.y.0 releases - skipping")
		return nil, nil
	}

	avg := sum / int64(len(deltas))

	var t float64
	avgDays := toDays(time.Duration(avg))
	for _, d := range deltas {
		distanceToAvg := toDays(d) - avgDays
		t += distanceToAvg * distanceToAvg
	}

	t /= float64(len(deltas))
	stddev := math.Sqrt(t)

	return &repoStats{
		owner:              owner,
		repo:               repo,
		dotZeroVersions:    len(deltas),
		minDaysBetween:     toDays(min),
		averageDaysBetween: avgDays,
		maxDaysBetween:     toDays(max),
		standardDeviation:  stddev,
	}, nil
}

func toDays(d time.Duration) float64 {
	return float64(d) / float64(24*time.Hour)
}

type release struct {
	Name        string
	PublishedAt time.Time
}

func fetchMajorReleases(ctx context.Context, client *githubv4.Client, owner, repo string) ([]release, error) {
	var q struct {
		Repository struct {
			Releases struct {
				Nodes    []release
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"releases(first:100, after:$after, orderBy:{direction:ASC, field:CREATED_AT})"`
		} `graphql:"repository(owner:$owner, name:$name)"`
	}

	inputs := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(repo),
		"after": (*githubv4.String)(nil),
	}

	var releases []release
	for {
		if err := client.Query(ctx, &q, inputs); err != nil {
			return nil, err
		}

		releases = append(releases, q.Repository.Releases.Nodes...)

		if !q.Repository.Releases.PageInfo.HasNextPage {
			break
		}

		inputs["after"] = githubv4.NewString(q.Repository.Releases.PageInfo.EndCursor)
	}

	return releases, nil
}
