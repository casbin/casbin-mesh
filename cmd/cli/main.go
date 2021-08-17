package main

import (
	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v2"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "show", Description: "Store the username and age"},
		{Text: "create", Description: "Store the article text posted by user"},
		{Text: "delete", Description: "Store the text commented to articles"},
		{Text: "exit"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursorWithSpace(), true)
}

func action(ctx *cli.Context) error {
	addr := ctx.String("Addr")
	c := NewCtx(NewClient(addr))
	p := prompt.New(c.Executor, c.Completer)
	p.Run()
	return nil
}

func main() {
<<<<<<< Updated upstream
<<<<<<< Updated upstream
	c := NewClient("localhost:4002")
=======
	c := NewClient("127.0.0.1:4004")
>>>>>>> Stashed changes
	ctx := NewCtx(c)
=======
	c := NewClient("127.0.0.1:4004")
	ctx := NewCtx(c)
	//err := ctx.LoadNamespaces()
	//if err != nil {
	//	fmt.Println(err)
	//}
>>>>>>> Stashed changes
	p := prompt.New(
		ctx.Executor,
		ctx.Completer,
		prompt.OptionCompletionOnDown(),
		prompt.OptionPrefix("127.0.0.1:4002 (Primary) >> "),
	)
	p.Run()
<<<<<<< Updated upstream
=======
	//app := &cli.App{
	//	Flags: []cli.Flag{
	//		&cli.StringFlag{
	//			Name:    "user",
	//			Aliases: []string{"u"},
	//			Usage:   "User",
	//		},
	//		&cli.StringFlag{
	//			Name:    "Addr",
	//			Aliases: []string{"addr"},
	//			Usage:   "Host",
	//		},
	//	},
	//	Action: action,
	//}
	//
	//err := app.Run(os.Args)
	//if err != nil {
	//	log.Fatal(err)
	//}
>>>>>>> Stashed changes
}
