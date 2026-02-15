package reviews

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/AndroidPoet/playconsole-cli/internal/api"
	"github.com/AndroidPoet/playconsole-cli/internal/cli"
	"github.com/AndroidPoet/playconsole-cli/internal/output"
)

var ReviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Manage app reviews",
	Long: `View and reply to user reviews on Google Play.

Reviews can be filtered and sorted to help you manage user feedback
and improve your app's rating.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List reviews",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific review",
	RunE:  runGet,
}

var replyCmd = &cobra.Command{
	Use:   "reply",
	Short: "Reply to a review",
	RunE:  runReply,
}

var (
	maxResults      int64
	startIndex      int64
	translationLang string
	reviewID        string
	replyText       string
	minRating       int
	maxRating       int
)

func init() {
	// List flags
	listCmd.Flags().Int64Var(&maxResults, "max-results", 50, "maximum number of results")
	listCmd.Flags().Int64Var(&startIndex, "start-index", 0, "starting index for pagination")
	listCmd.Flags().StringVar(&translationLang, "translation-lang", "", "translate reviews to this language")
	listCmd.Flags().IntVar(&minRating, "min-rating", 0, "minimum star rating (1-5)")
	listCmd.Flags().IntVar(&maxRating, "max-rating", 0, "maximum star rating (1-5)")

	// Get flags
	getCmd.Flags().StringVar(&reviewID, "review-id", "", "review ID")
	getCmd.Flags().StringVar(&translationLang, "translation-lang", "", "translate to this language")
	getCmd.MarkFlagRequired("review-id")

	// Reply flags
	replyCmd.Flags().StringVar(&reviewID, "review-id", "", "review ID")
	replyCmd.Flags().StringVar(&replyText, "text", "", "reply text")
	replyCmd.MarkFlagRequired("review-id")
	replyCmd.MarkFlagRequired("text")

	ReviewsCmd.AddCommand(listCmd)
	ReviewsCmd.AddCommand(getCmd)
	ReviewsCmd.AddCommand(replyCmd)
}

// ReviewInfo represents review information
type ReviewInfo struct {
	ReviewID     string `json:"review_id"`
	AuthorName   string `json:"author_name"`
	Rating       int64  `json:"rating"`
	Text         string `json:"text"`
	LastModified string `json:"last_modified,omitempty"`
	AppVersion   string `json:"app_version,omitempty"`
	DeviceType   string `json:"device_type,omitempty"`
	HasReply     bool   `json:"has_reply"`
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	call := client.Reviews().List(client.GetPackageName()).Context(ctx)

	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	if startIndex > 0 {
		call = call.StartIndex(startIndex)
	}
	if translationLang != "" {
		call = call.TranslationLanguage(translationLang)
	}

	reviews, err := call.Do()
	if err != nil {
		return err
	}

	result := make([]ReviewInfo, 0)
	for _, r := range reviews.Reviews {
		// Get the user's comment (first comment in the thread)
		if len(r.Comments) == 0 {
			continue
		}

		userComment := r.Comments[0].UserComment
		if userComment == nil {
			continue
		}

		// Filter by rating if specified
		if minRating > 0 && int(userComment.StarRating) < minRating {
			continue
		}
		if maxRating > 0 && int(userComment.StarRating) > maxRating {
			continue
		}

		info := ReviewInfo{
			ReviewID:   r.ReviewId,
			AuthorName: r.AuthorName,
			Rating:     int64(userComment.StarRating),
			Text:       userComment.Text,
			HasReply:   len(r.Comments) > 1 && r.Comments[1].DeveloperComment != nil,
		}

		if userComment.LastModified != nil {
			info.LastModified = time.Unix(userComment.LastModified.Seconds, 0).Format(time.RFC3339)
		}
		if userComment.AppVersionCode > 0 {
			info.AppVersion = fmt.Sprintf("%d", userComment.AppVersionCode)
		}
		if userComment.Device != "" {
			info.DeviceType = userComment.Device
		}

		result = append(result, info)
	}

	if len(result) == 0 {
		output.PrintInfo("No reviews found")
		return nil
	}

	return output.Print(result)
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	call := client.Reviews().Get(client.GetPackageName(), reviewID).Context(ctx)
	if translationLang != "" {
		call = call.TranslationLanguage(translationLang)
	}

	review, err := call.Do()
	if err != nil {
		return err
	}

	result := map[string]interface{}{
		"review_id":   review.ReviewId,
		"author_name": review.AuthorName,
	}

	if len(review.Comments) > 0 && review.Comments[0].UserComment != nil {
		uc := review.Comments[0].UserComment
		result["rating"] = uc.StarRating
		result["text"] = uc.Text
		result["language"] = uc.ReviewerLanguage
		result["device"] = uc.Device
		result["app_version_code"] = uc.AppVersionCode
		result["app_version_name"] = uc.AppVersionName

		if uc.LastModified != nil {
			result["last_modified"] = time.Unix(uc.LastModified.Seconds, 0).Format(time.RFC3339)
		}
	}

	// Check for developer reply
	if len(review.Comments) > 1 && review.Comments[1].DeveloperComment != nil {
		dc := review.Comments[1].DeveloperComment
		result["reply"] = map[string]interface{}{
			"text": dc.Text,
			"last_modified": func() string {
				if dc.LastModified != nil {
					return time.Unix(dc.LastModified.Seconds, 0).Format(time.RFC3339)
				}
				return ""
			}(),
		}
	}

	return output.Print(result)
}

func runReply(cmd *cobra.Command, args []string) error {
	if err := cli.RequirePackage(cmd); err != nil {
		return err
	}

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would reply to review %s", reviewID)
		return nil
	}

	client, err := api.NewClient(cli.GetPackageName(), 60*time.Second)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	replyReq := &androidpublisher.ReviewsReplyRequest{
		ReplyText: replyText,
	}

	reply, err := client.Reviews().Reply(client.GetPackageName(), reviewID, replyReq).Context(ctx).Do()
	if err != nil {
		return err
	}

	output.PrintSuccess("Reply posted to review %s", reviewID)

	if reply.Result != nil && reply.Result.ReplyText != "" {
		return output.Print(map[string]interface{}{
			"review_id":  reviewID,
			"reply_text": reply.Result.ReplyText,
			"last_edited": func() string {
				if reply.Result.LastEdited != nil {
					return time.Unix(reply.Result.LastEdited.Seconds, 0).Format(time.RFC3339)
				}
				return ""
			}(),
		})
	}

	return nil
}
