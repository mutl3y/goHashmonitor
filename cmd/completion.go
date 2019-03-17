package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// completionCmd represents the completion command
// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:    "completion",
	Hidden: true,
	Short:  "Generates bash completion scripts",
	Long: `To load completion run

. <(bitbucket completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(bitbucket completion)
`,
	Run: func(cmd *cobra.Command, args []string) {

		f, err := os.OpenFile("addThisToBashRC.txt", os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = f.Close()
		if err != nil {
			log.Fatalf("failed to close file %v", f.Name())
		}
		err = rootCmd.GenBashCompletion(f)
		if err != nil {
			log.Fatalf("failed to generate bash completions")
		}
		fmt.Printf("Add the contents of addThisToBashRC.txt to either your .bashrc or .profile ")
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
