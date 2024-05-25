// Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/chzyer/readline"
	"github.com/olekukonko/tablewriter"
)

var (
	mu      sync.Mutex
	arrCmds []baseCommandAction
	idxCmds = map[string]baseCommandAction{}
)

type baseCommandSpec struct {
	Mgr  bool
	Path string
	Desc string
}

type baseCommandAction interface {
	Spec() baseCommandSpec
	Action(fg flagSet, l *readline.Instance) (string, error)
}

func init() {
	register(new(cmdHelp))
}

func register(fn baseCommandAction) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := idxCmds[fn.Spec().Path]; !ok {
		idxCmds[fn.Spec().Path] = fn
		arrCmds = append(arrCmds, fn)

		sort.Slice(arrCmds, func(i, j int) bool {
			return strings.Compare(arrCmds[i].Spec().Path, arrCmds[j].Spec().Path) < 0
		})
	}
}

type cmdHelp struct{}

func (cmdHelp) Spec() baseCommandSpec {
	return baseCommandSpec{
		Path: "help",
	}
}

func (cmdHelp) Action(fg flagSet, l *readline.Instance) (string, error) {

	var tbuf bytes.Buffer

	// base commands
	{
		tbuf.WriteString("Common Commands:\n")
		var (
			tableBase = tablewriter.NewWriter(&tbuf)
		)

		tableBase.SetRowLine(false)
		tableBase.SetColumnSeparator("")
		tableBase.SetHeaderLine(false)
		tableBase.SetBorder(false)
		tableBase.EnableBorder(false)
		tableBase.SetAutoWrapText(false)
		tableBase.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tableBase.SetAlignment(tablewriter.ALIGN_LEFT)

		// table.SetHeader([]string{"Command", "Usage"})

		for _, c := range arrCmds {
			if c.Spec().Mgr || c.Spec().Path == "help" {
				continue
			}
			tableBase.Append([]string{c.Spec().Path, c.Spec().Desc})
		}

		tableBase.Append([]string{"help", ""})
		// table.Append([]string{"quit", ""})

		tableBase.Render()
	}

	// management commands
	{
		tbuf.WriteString("\nManagement Commands:\n")
		var (
			tableMgr = tablewriter.NewWriter(&tbuf)
		)

		tableMgr.SetRowLine(false)
		tableMgr.SetColumnSeparator("")
		tableMgr.SetHeaderLine(false)
		tableMgr.SetBorder(false)
		tableMgr.EnableBorder(false)
		tableMgr.SetAutoWrapText(false)
		tableMgr.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tableMgr.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, c := range arrMgrServices {
			tableMgr.Append([]string{c.lowerName, c.name})
		}

		tableMgr.Render()
	}

	return fmt.Sprintf("\n%s\n", tbuf.String()), nil
}

func cliInvoke(s string, l *readline.Instance) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	fg := flagParse(s)

	for _, c := range arrCmds {
		if !strings.HasPrefix(fg.path, c.Spec().Path) {
			continue
		}

		if fg.path != c.Spec().Path {
			fg.path = strings.TrimSpace(fg.path[len(c.Spec().Path):])
		} else {
			fg.path = ""
		}

		fg.varArgs = flagVarParse(fg.path)

		return c.Action(fg, l)
	}

	return "", fmt.Errorf("no command match")
}
