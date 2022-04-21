package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pacevedom/kubectl-cvo/pkg/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Removes an override in CVO (if it exists) to manage an operator",
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Println("Error when building client config:", err)
			return
		}
		cl, err := client.NewClient(config)
		if err != nil {
			fmt.Println("Error when building client:", err)
			return
		}
		operators, err := cl.ListUnmanagedOperators()
		if err != nil {
			fmt.Println("Error when fetching unmanaged operators:", err)
			return
		}
		if len(operators) == 0 {
			fmt.Println("No unmanaged operators found")
			return
		}
		var qs = []*survey.Question{
			{
				Name: "operator",
				Prompt: &survey.Select{
					Message: "Choose an operator:",
					Options: operators,
				},
			},
		}
		answer := struct {
			Operator string `survey:"operator"`
		}{}

		err = survey.Ask(qs, &answer)
		if err != nil {
			fmt.Println("Error when getting answer from selector:", err)
			return
		}
		err = cl.ManageOperator(answer.Operator)
		if err != nil {
			fmt.Println("Error when managing operator: %w", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(manageCmd)
}
