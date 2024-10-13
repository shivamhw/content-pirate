package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type helloCfg struct{
	name string
	greeting string
}

var cfg helloCfg
var helloCmd = &cobra.Command{
	Use:   "hello-world",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: helloCmdInit,
}

type helloClient struct {
	name	string
	greeting	string
}

func NewHello(cfg *helloCfg) *helloClient {
	return &helloClient{
		name: cfg.name,
		greeting: cfg.greeting,
	}
}

func (h *helloClient) Run(){
	fmt.Println(h.greeting + " " + h.name)
}

func init(){
	helloCmd.Flags().StringVar(&cfg.name, "name", "default", "pass the name")
	helloCmd.Flags().StringVar(&cfg.greeting, "greet", "default", "pass the name")
}

func helloCmdInit(cmd *cobra.Command, args []string){
	fmt.Println("this is init of hello world")
	client := NewHello(&cfg)
	client.Run()

}