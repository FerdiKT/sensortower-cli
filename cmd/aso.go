package cmd

import (
	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/ferdikt/sensortower-cli/internal/textutil"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(asoCmd)
	asoCmd.AddCommand(asoMetadataAuditCmd, asoKeywordGapCmd)
	asoMetadataAuditCmd.Flags().Int64("app-id", 0, "App ID")
	asoMetadataAuditCmd.Flags().String("country", "US", "Store country")
	_ = asoMetadataAuditCmd.MarkFlagRequired("app-id")
	asoKeywordGapCmd.Flags().Int64("app-id", 0, "Target app ID")
	asoKeywordGapCmd.Flags().String("country", "US", "Store country")
	asoKeywordGapCmd.Flags().String("competitor-ids-file", "", "Competitor app IDs file")
	_ = asoKeywordGapCmd.MarkFlagRequired("app-id")
	_ = asoKeywordGapCmd.MarkFlagRequired("competitor-ids-file")
}

var asoCmd = &cobra.Command{Use: "aso", Short: "ASO-oriented helper commands"}

var asoMetadataAuditCmd = &cobra.Command{
	Use:   "metadata-audit",
	Short: "Audit app metadata for ASO completeness",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		appID, _ := cmd.Flags().GetInt64("app-id")
		country, _ := cmd.Flags().GetString("country")
		resp, _, err := client.AppDetails(commandContext(cmd), appID, country)
		if err != nil {
			return err
		}
		return writeOutput(structToMap(auditMetadata(resp)))
	},
}

var asoKeywordGapCmd = &cobra.Command{
	Use:   "keyword-gap",
	Short: "Find competitor terms absent from target app metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		appID, _ := cmd.Flags().GetInt64("app-id")
		country, _ := cmd.Flags().GetString("country")
		file, _ := cmd.Flags().GetString("competitor-ids-file")
		competitorIDs, err := readInt64Lines(file)
		if err != nil {
			return clierror.Wrap(11, err.Error())
		}
		target, _, err := client.AppDetails(commandContext(cmd), appID, country)
		if err != nil {
			return err
		}
		targetTerms := appKeywords(target)
		competitorTerms := map[string]int{}
		for _, competitorID := range competitorIDs {
			resp, _, err := client.AppDetails(commandContext(cmd), competitorID, country)
			if err != nil {
				continue
			}
			for word, score := range appKeywords(resp) {
				competitorTerms[word] += score
			}
		}
		return writeOutput(structToMap(sensortower.KeywordGapResult{
			TargetAppID:      appID,
			CompetitorAppIDs: competitorIDs,
			MissingKeywords:  textutil.SortedDiff(targetTerms, competitorTerms, 30),
		}))
	},
}

func auditMetadata(app *sensortower.AppDetails) sensortower.MetadataAudit {
	var issues []string
	if app.Subtitle == "" {
		issues = append(issues, "missing subtitle")
	}
	if len(app.Description.FullDescription) < 300 {
		issues = append(issues, "description is short")
	}
	if app.PromoText == "" {
		issues = append(issues, "missing promo_text")
	}
	if len(app.SupportedLanguages) < 2 {
		issues = append(issues, "only one supported language")
	}
	return sensortower.MetadataAudit{
		AppID:    app.AppID,
		Name:     app.Name,
		Issues:   issues,
		Keywords: textutil.SortedDiff(map[string]int{}, appKeywords(app), 20),
	}
}

func appKeywords(app *sensortower.AppDetails) map[string]int {
	return textutil.Keywords(app.Name, app.Subtitle, app.Description.FullDescription)
}
