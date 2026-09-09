package diff

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

// DiffCmd compares edit state vs live version
var DiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between draft edit and live app",
	Long: `Compare a draft edit against the live version of the app.

The live state is read through a throwaway edit that is discarded
afterwards; the draft edit given with --edit-id is left untouched.
Listings are keyed by language and tracks by track name. Without
--edit-id there is no draft to compare, so all sets are empty.`,
	RunE: runDiff,
}

var (
	editID  string
	section string
)

func init() {
	DiffCmd.Flags().StringVar(&editID, "edit-id", "", "existing edit ID to compare against live (optional)")
	DiffCmd.Flags().StringVar(&section, "section", "all", "section to diff: listings, tracks, all")
}

// DiffResult represents the diff of one section
type DiffResult struct {
	Section   string   `json:"section"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Changed   []string `json:"changed"`
	Unchanged int      `json:"unchanged"`
	Error     string   `json:"error,omitempty"`
}

func newDiffResult(section string) DiffResult {
	return DiffResult{
		Section: section,
		Added:   []string{},
		Removed: []string{},
		Changed: []string{},
	}
}

// compare fills result from two keyed maps: keys only in draft are added,
// keys only in base are removed, and shared keys are changed or unchanged
// according to equal.
func compare[T any](result *DiffResult, base, draft map[string]T, equal func(a, b T) bool) {
	for key, d := range draft {
		b, ok := base[key]
		switch {
		case !ok:
			result.Added = append(result.Added, key)
		case equal(b, d):
			result.Unchanged++
		default:
			result.Changed = append(result.Changed, key)
		}
	}
	for key := range base {
		if _, ok := draft[key]; !ok {
			result.Removed = append(result.Removed, key)
		}
	}
	sort.Strings(result.Added)
	sort.Strings(result.Removed)
	sort.Strings(result.Changed)
}

func fetchListings(client *api.Client, edit *api.Edit) (map[string]*androidpublisher.Listing, error) {
	resp, err := edit.Listings().List(client.GetPackageName(), edit.ID()).Context(edit.Context()).Do()
	if err != nil {
		return nil, err
	}
	listings := make(map[string]*androidpublisher.Listing, len(resp.Listings))
	for _, l := range resp.Listings {
		listings[l.Language] = l
	}
	return listings, nil
}

func listingsEqual(a, b *androidpublisher.Listing) bool {
	return a.Title == b.Title &&
		a.ShortDescription == b.ShortDescription &&
		a.FullDescription == b.FullDescription &&
		a.Video == b.Video
}

func fetchTracks(client *api.Client, edit *api.Edit) (map[string]*androidpublisher.Track, error) {
	resp, err := edit.Tracks().List(client.GetPackageName(), edit.ID()).Context(edit.Context()).Do()
	if err != nil {
		return nil, err
	}
	tracks := make(map[string]*androidpublisher.Track, len(resp.Tracks))
	for _, t := range resp.Tracks {
		tracks[t.Track] = t
	}
	return tracks, nil
}

// releaseSignature summarises a track's releases independent of order
func releaseSignature(t *androidpublisher.Track) string {
	sigs := make([]string, 0, len(t.Releases))
	for _, r := range t.Releases {
		sigs = append(sigs, fmt.Sprintf("%s|%s|%g|%v", r.Name, r.Status, r.UserFraction, []int64(r.VersionCodes)))
	}
	sort.Strings(sigs)
	return strings.Join(sigs, "\n")
}

func tracksEqual(a, b *androidpublisher.Track) bool {
	return releaseSignature(a) == releaseSignature(b)
}

func diffListings(client *api.Client, baseline, draft *api.Edit) DiffResult {
	result := newDiffResult("listings")

	base, err := fetchListings(client, baseline)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	current := base
	if draft != baseline {
		current, err = fetchListings(client, draft)
		if err != nil {
			result.Error = err.Error()
			return result
		}
	}

	compare(&result, base, current, listingsEqual)
	return result
}

func diffTracks(client *api.Client, baseline, draft *api.Edit) DiffResult {
	result := newDiffResult("tracks")

	base, err := fetchTracks(client, baseline)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	current := base
	if draft != baseline {
		current, err = fetchTracks(client, draft)
		if err != nil {
			result.Error = err.Error()
			return result
		}
	}

	compare(&result, base, current, tracksEqual)
	return result
}

func runDiff(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	switch section {
	case "all", "listings", "tracks":
	default:
		return fmt.Errorf("invalid section '%s': use listings, tracks or all", section)
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	// A fresh edit reflects the live app; it is discarded on Close.
	baseline, err := client.CreateEdit()
	if err != nil {
		return err
	}
	defer baseline.Close()

	// The draft edit belongs to the caller and is kept on Close.
	draft := baseline
	if editID != "" {
		draft, err = client.GetEdit(editID)
		if err != nil {
			return err
		}
		defer draft.Close()
	} else {
		output.PrintInfo("No --edit-id given: a fresh edit equals the live app, so there is nothing to compare")
	}

	results := make([]DiffResult, 0, 2)

	if section == "all" || section == "listings" {
		results = append(results, diffListings(client, baseline, draft))
	}

	if section == "all" || section == "tracks" {
		results = append(results, diffTracks(client, baseline, draft))
	}

	return output.Print(results)
}
