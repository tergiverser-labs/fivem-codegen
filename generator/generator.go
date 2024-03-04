package generator

import (
	"encoding/json"
	"fivem-codegen/utils"
	"fmt"
	"os"
)

type nativesDb map[string]map[string]struct {
	Name string `json:"name"`
}

type Declaration struct {
	Side        string
	Hash        string
	Name        string
	Args        string
	TsArgs      string
	ReturnValue string
	PreCall     string
	Call        string
	PostCall    string
}

type Generator struct {
	FileClient *os.File
	FileServer *os.File

	Fps          []string
	nativesDb    *nativesDb
	Declarations []Declaration
}

func InitGenerator() (*Generator, error) {
	fClient, err := os.OpenFile("client.ts", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	fServer, err := os.OpenFile("server.ts", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile("deps/native-db/natives.json")
	if err != nil {
		return nil, err
	}

	nativesDb := &nativesDb{}
	if err := json.Unmarshal(bytes, &nativesDb); err != nil {
		return nil, err
	}

	return &Generator{
		FileClient: fClient,
		FileServer: fServer,

		Fps:          make([]string, 0),
		nativesDb:    nativesDb,
		Declarations: make([]Declaration, 0),
	}, nil
}

func (g *Generator) GetRealNativeName(nativeName string) *string {
	for _, vCategoryMap := range *g.nativesDb {
		for kName, dbNative := range vCategoryMap {
			if kName == nativeName {
				return &dbNative.Name
			}
		}
	}

	return nil
}

func (g *Generator) computeNativeName(nativeName string) string {
	if realName := g.GetRealNativeName(nativeName[1:]); realName != nil {
		return utils.SnakeToCamelCase(*realName)
	}

	if nativeName[0:3] == "_0x" {
		return "N" + nativeName
	}

	return utils.SnakeToCamelCase(nativeName)
}

func (g *Generator) WriteDeclarations() error {
	for _, declaration := range g.Declarations {
		banner := ""
		if declaration.Hash == "" {
			banner = "/*\n * @fivem custom native\n */\n"
		}

		line := fmt.Sprintf("%sexport function %s(%s): %s {\n    // @ts-expect-error fwdecc\n    %s%s%s\n}\n",
			banner,
			g.computeNativeName(declaration.Name),
			declaration.TsArgs,
			declaration.ReturnValue,
			declaration.PreCall,
			declaration.Call,
			declaration.PostCall,
		)

		switch declaration.Side {
		case "shared":
			{
				if _, err := fmt.Fprintln(g.FileClient, line); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(g.FileServer, line); err != nil {
					return err
				}

				break
			}
		case "client":
			{
				if _, err := fmt.Fprintln(g.FileClient, line); err != nil {
					return err
				}

				break
			}
		case "server":
			{
				if _, err := fmt.Fprintln(g.FileServer, line); err != nil {
					return err
				}

				break
			}
		}
	}

	return nil
}
