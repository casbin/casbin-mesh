/*
@Author: Weny Xu
@Date: 2021/08/09 19:41
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/tidwall/pretty"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Context interface {
	ExecutorFn
	CompleterFn
}

type ExecutorFn interface {
	Executor(string)
}

type CompleterFn interface {
	Completer(prompt.Document) []prompt.Suggest
}

var TXSuggests = []prompt.Suggest{
	{"abort", "Abort"},
	{"commit", "Commit"},
	{"delete", "Delete"},
}
var AddSuggests = []prompt.Suggest{
	{"add", "Add"},
}
var RemoveSuggests = []prompt.Suggest{
	{"remove", "Remove"},
}
var UpdateSuggests = []prompt.Suggest{
	{"update", "Update"},
}

var NamespaceSuggests = append([]prompt.Suggest{
	{"print model", "Print model"},
	{"show policies", "List all policies"},
}, TopLevelSuggests...)

var TopLevelSuggests = merge([]prompt.Suggest{
	{"stats", "Show statics"},
	{"use", "Set namespace"},
	{"exit", "Exit"},
	{"show namespaces", "Show namespaces"},
}, AddSuggests, UpdateSuggests, RemoveSuggests)

func merge(ps ...[]prompt.Suggest) []prompt.Suggest {
	var out []prompt.Suggest
	for _, p := range ps {
		out = append(out, p...)
	}
	return out
}

func updateOperationsRemove(slice [][]CasbinRule, s int) [][]CasbinRule {
	return append(slice[:s], slice[s+1:]...)
}

type ctx struct {
	*client
	transaction string
	//can be "g" or "p"
	sec              string
	ptype            string
	operations       Rules
	updateOperations RuleGroups
	cmd              string
	namespace        string
	namespaces       []prompt.Suggest
}

func (c *ctx) LoadNamespaces() error {
	ns, err := c.ListNamespaces(context.TODO())
	if err != nil {
		return err
	}
	for _, s := range ns {
		c.namespaces = append(c.namespaces, prompt.Suggest{Text: s, Description: "Namespace"})
	}
	return nil
}

func NewCtx(c *client) *ctx {
	return &ctx{
		client: c,
	}
}

func (c *ctx) PrintUpdateOperations() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"INDEX", "Type", "V0", "V1", "V2", "V3", "V4", "V5", "->", "Type", "V0", "V1", "V2", "V3", "V4", "V5"})
	for index, rule := range c.updateOperations {
		if len(rule) > 1 {
			t.AppendRow(table.Row{index,
				rule[0].PType, rule[0].V0, rule[0].V1, rule[0].V2, rule[0].V3, rule[0].V4, rule[0].V5, "",
				rule[1].PType, rule[1].V0, rule[1].V1, rule[1].V2, rule[1].V3, rule[1].V4, rule[1].V5,
			})
		} else {
			t.AppendRow(table.Row{index,
				rule[0].PType, rule[0].V0, rule[0].V1, rule[0].V2, rule[0].V3, rule[0].V4, rule[0].V5, "",
			})
		}

	}
	t.Render()
}

func (c *ctx) PrintOperations() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"INDEX", "Type", "V0", "V1", "V2", "V3", "V4", "V5"})
	for index, rule := range c.operations {
		t.AppendRow(table.Row{index, rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5})
	}
	t.Render()
}

var secToString = map[string]string{"p": "Policies", "g": "Groups"}

func (c *ctx) Uncommitted() {
	fmt.Printf("Has a commited %s%s transaction. commit or abort it\n", c.transaction, secToString[c.sec])
}

func (c *ctx) HandleTXInput(name string, argv []string) {
	if c.transaction != name && c.transaction != "" {
		c.Uncommitted()
		return
	}
	c.transaction = name
	var (
		sec   string
		ptype string
	)
	if len(argv) > 0 {
		ptype = argv[0]
		if argv[0] != "" {
			sec = string(argv[0][0])
			if sec != "g" && sec != "p" {
				fmt.Printf("invalid sec type\n")
				return
			}
		}
		if c.ptype != "" && c.ptype != ptype {
			c.Uncommitted()
			return
		}
		c.sec = sec
		c.ptype = ptype
		c.operations = append(c.operations, convertRule(ptype, argv[1:]))
		fmt.Printf("Uncommitted %s%s transaction\n", c.transaction, secToString[c.sec])
		c.PrintOperations()
	}
}

func (c *ctx) HandleUpdateInput(name string, argv []string) {
	if c.transaction != name && c.transaction != "" {
		c.Uncommitted()
		return
	}
	c.transaction = name
	var (
		sec   string
		ptype string
	)
	if len(argv) > 0 {
		ptype = argv[0]
		if argv[0] != "" {
			sec = string(argv[0][0])
			if sec != "g" && sec != "p" {
				fmt.Printf("invalid sec type\n")
				return
			}
		}
		if c.sec != "" && c.sec != sec {
			c.Uncommitted()
			return
		}
		c.sec = sec
		if len(c.updateOperations) > 0 {
			last := &c.updateOperations[len(c.updateOperations)-1]
			if len(*last) == 1 {
				*last = append(*last, convertRule(ptype, argv[1:]))
			} else {
				c.updateOperations = append(c.updateOperations, append([]CasbinRule{}, convertRule(ptype, argv[1:])))
			}
		} else {
			c.updateOperations = append(c.updateOperations, append([]CasbinRule{}, convertRule(ptype, argv[1:])))
		}

		fmt.Printf("Uncommitted %s%s transaction\n", c.transaction, secToString[c.sec])
		c.PrintUpdateOperations()
	}
}
func (c *ctx) InitTx() {
	c.transaction = ""
	c.sec = ""
	c.operations = nil
	c.updateOperations = nil
}
func (c *ctx) Executor(line string) {
	cmds := strings.Split(strings.TrimSpace(line), " ")
	var argv []string
	if len(cmds) > 0 {
		c.cmd = cmds[0]
		argv = cmds[1:]
	}
	switch strings.ToUpper(cmds[0]) {
	case "DELETE":
		if len(argv) > 0 {
			index, err := strconv.Atoi(argv[0])
			if err != nil {
				fmt.Println(err)
			}
			if index > -1 && index < len(c.updateOperations) {
				c.updateOperations = updateOperationsRemove(c.updateOperations, index)
				fmt.Printf("Deleted item of index %d\n", index)
			}
		}
	case "STATS":
		stats, err := c.ShowStats(context.TODO())
		if err != nil {
			fmt.Printf("Error:%s", err.Error())
		}
		fmt.Println(string(pretty.Color(pretty.Pretty(stats), nil)))
	case "ABORT":
		c.updateOperations = nil
		c.transaction = ""
		c.sec = ""
		c.operations = nil
		fmt.Printf("Transaction aborted!\n")
	case "ADD":
		c.HandleTXInput("add", argv)
	case "REMOVE":
		c.HandleTXInput("remove", argv)
	case "UPDATE":
		c.HandleUpdateInput("update", argv)
	case "SET":
	case "COMMIT":
		start := time.Now()
		ns := c.namespace
		sec := c.sec
		ptype := c.ptype
		switch c.transaction {
		case "add":
			policies, err := c.client.AddPolicies(context.TODO(), ns, sec, ptype, c.operations.to2DString())
			if err != nil {
				fmt.Printf("Commit failed:%s", err.Error())
				return
			}
			var tmp Rules
			for _, p := range policies {
				tmp = append(tmp, convertRule(ptype, p))
			}
			elapsed := time.Since(start)
			if tmp != nil {
				fmt.Printf("%d rules effected <%s>\n", len(tmp), elapsed)
				tmp.Print()
				c.InitTx()
			} else {
				fmt.Printf("No effected rules! <%s>\n", elapsed)
			}
		case "update":
			o, n := c.updateOperations.to2DString()
			effected, err := c.client.UpdatePolicies(context.TODO(), ns, sec, ptype, o, n)
			if err != nil {
				fmt.Printf("Commit failed:%s", err.Error())
				return
			}
			elapsed := time.Since(start)
			if effected {
				fmt.Printf("Effected <%s>\n", elapsed)
				c.InitTx()
			} else {
				fmt.Printf("No effected rules! <%s>\n", elapsed)
			}
		case "remove":
			policies, err := c.client.RemovePolicies(context.TODO(), ns, sec, ptype, c.operations.to2DString())
			if err != nil {
				fmt.Printf("Commit failed:%s", err.Error())
				return
			}
			var tmp Rules
			for _, p := range policies {
				tmp = append(tmp, convertRule(ptype, p))
			}
			elapsed := time.Since(start)
			if tmp != nil {
				fmt.Printf("%d rules effected <%s>\n", len(tmp), elapsed)
				tmp.Print()
				c.InitTx()
			} else {
				fmt.Printf("No effected rules! <%s>\n", elapsed)
			}
		}
	case "USE":
		if len(argv) > 0 {
			c.namespace = argv[0]
			fmt.Printf("Set namespace: %s\n", argv[0])
		}
	case "PRINT":
		if len(argv) > 0 {
			if argv[0] == "model" {
				model, err := c.PrintModel(context.Background(), c.namespace)
				if err != nil {
					log.Printf("Error:%s\n", err.Error())
				}
				fmt.Printf(model)
			}
		}
	case "SHOW":
		if len(argv) > 0 {
			switch strings.ToUpper(argv[0]) {
			case "POLICIES":
				ns, err := c.ListPolicies(context.TODO(), c.namespace)
				if err != nil {
					log.Printf("Error:%s\n", err.Error())
				}
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"Key", "Type", "V0", "V1", "V2", "V3", "V4", "V5"})
				for _, s := range ns {
					rule := CasbinRule{}
					err := json.Unmarshal([]byte(s[1]), &rule)
					if err != nil {
						fmt.Printf("Error:%s\n", err.Error())
					}
					t.AppendRow(table.Row{s[0], rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5})
				}
				t.Render()
			case "NAMESPACES":
				c.namespaces = nil
				ns, err := c.ListNamespaces(context.TODO())
				if err != nil {
					log.Printf("Error:%s\n", err.Error())
				}
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"Namespace"})
				for _, s := range ns {
					c.namespaces = append(c.namespaces, prompt.Suggest{Text: s, Description: "Namespace"})
					t.AppendRow(table.Row{s})
				}
				t.Render()
			}
		}
	case "QUIT", "EXIT":
		fmt.Println("Bye~")
		os.Exit(0)
	}
}

func (c *ctx) Completer(d prompt.Document) []prompt.Suggest {
	var bound []prompt.Suggest
	if c.transaction == "" {
		if c.namespace != "" {
			bound = append(NamespaceSuggests, c.namespaces...)
		} else {
			bound = append(TopLevelSuggests, c.namespaces...)
		}
	}
	switch c.transaction {
	case "add":
		bound = append(append(AddSuggests, TXSuggests...), bound...)
	case "remove":
		bound = append(append(RemoveSuggests, TXSuggests...), bound...)
	case "update":
		bound = append(append(UpdateSuggests, TXSuggests...), bound...)
		if len(c.updateOperations) > 0 {
			last := c.updateOperations[len(c.updateOperations)-1]
			if len(last) == 1 {
				rule := last[0]
				bound = append([]prompt.Suggest{{
					Text: strings.TrimSpace(
						fmt.Sprintf("update %s %s %s %s %s %s %s", rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5)),
				},
				}, bound...)
			}
		}
	}
	return prompt.FilterHasPrefix(bound, d.GetWordBeforeCursorWithSpace(), true)

}
