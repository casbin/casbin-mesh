package main

import (
	"github.com/c-bata/go-prompt"
)

//func completer(d prompt.Document) []prompt.Suggest {
//	s := []prompt.Suggest{
//		{Text: "show", Description: "Store the username and age"},
//		{Text: "create", Description: "Store the article text posted by user"},
//		{Text: "delete", Description: "Store the text commented to articles"},
//		{Text: "exit"},
//	}
//	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursorWithSpace(), true)
//}
//
//func action(ctx *cli.Context) error {
//	addr := ctx.String("Addr")
//	c := NewCtx(NewClient(options{}))
//	p := prompt.New(c.Executor, c.Completer)
//	p.Run()
//	return nil
//}

func main() {
	c := NewClient(options{
		target:   "localhost:4002",
		authType: Basic,
		username: "root",
		password: "root",
	})
	ctx := NewCtx(c)
	p := prompt.New(
		ctx.Executor,
		ctx.Completer,
		prompt.OptionCompletionOnDown(),
		prompt.OptionPrefix("127.0.0.1:4002 (Primary) >> "),
	)
	p.Run()
}
