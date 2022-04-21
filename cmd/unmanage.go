package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pacevedom/kubectl-cvo/pkg/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var unmanageCmd = &cobra.Command{
	Use:   "unmanage",
	Short: "Sets an override in CVO to unmanage an operator",
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
		operators, err := cl.ListManagedOperators()
		if err != nil {
			fmt.Println("Error when fetching managed operators:", err)
			return
		}
		if len(operators) == 0 {
			fmt.Println("No managed operators found")
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
		err = cl.UnmanageOperator(answer.Operator)
		if err != nil {
			fmt.Println("Error when unmanaging operator:", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(unmanageCmd)
}
