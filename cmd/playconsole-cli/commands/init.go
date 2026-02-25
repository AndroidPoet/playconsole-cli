package commands

import (
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/apks"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/apps"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/auth"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/availability"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/bundles"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/completion"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/deobfuscation"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/devices"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/devicetiers"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/diff"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/doctor"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/edits"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/externaltx"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/images"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/initcmd"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/listings"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/offers"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/orders"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/products"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/purchases"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/recovery"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/reports"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/reviews"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/setup"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/stats"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/subscriptions"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/testing"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/tracks"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/users"
	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands/vitals"
)

func init() {
	// Add all command groups to root

	// Core app management
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(apps.AppsCmd)
	rootCmd.AddCommand(tracks.TracksCmd)
	rootCmd.AddCommand(bundles.BundlesCmd)
	rootCmd.AddCommand(apks.APKsCmd)
	rootCmd.AddCommand(listings.ListingsCmd)
	rootCmd.AddCommand(images.ImagesCmd)
	rootCmd.AddCommand(reviews.ReviewsCmd)
	rootCmd.AddCommand(products.ProductsCmd)
	rootCmd.AddCommand(subscriptions.SubscriptionsCmd)
	rootCmd.AddCommand(purchases.PurchasesCmd)
	rootCmd.AddCommand(edits.EditsCmd)
	rootCmd.AddCommand(users.UsersCmd)
	rootCmd.AddCommand(testing.TestingCmd)
	rootCmd.AddCommand(setup.SetupCmd)
	rootCmd.AddCommand(vitals.VitalsCmd)
	rootCmd.AddCommand(devices.DevicesCmd)
	rootCmd.AddCommand(reports.ReportsCmd)
	rootCmd.AddCommand(stats.StatsCmd)

	// New commands
	rootCmd.AddCommand(completion.CompletionCmd)
	rootCmd.AddCommand(doctor.DoctorCmd)
	rootCmd.AddCommand(initcmd.InitCmd)
	rootCmd.AddCommand(offers.OffersCmd)
	rootCmd.AddCommand(deobfuscation.DeobfuscationCmd)
	rootCmd.AddCommand(orders.OrdersCmd)
	rootCmd.AddCommand(externaltx.ExternalTxCmd)
	rootCmd.AddCommand(recovery.RecoveryCmd)
	rootCmd.AddCommand(devicetiers.DeviceTiersCmd)
	rootCmd.AddCommand(availability.AvailabilityCmd)
	rootCmd.AddCommand(diff.DiffCmd)
}
