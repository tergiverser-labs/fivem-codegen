package main

import (
	"bufio"
	"fivem-codegen/generator"
	"fivem-codegen/utils"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func computeCfxNativeName(nativeName string) string {
	if nativeName[0:3] == "_0x" {
		return "N" + strings.ToLower(nativeName)
	}

	return utils.FirstToUpper(utils.SnakeToCamelCase(nativeName))
}

func parseGtaNative(fp string) (*generator.Declaration, error) {
	f, err := os.OpenFile(fp, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	declaration := &generator.Declaration{Side: "client"}
	isInCodeBlock := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "game:") && !strings.Contains(line, "gta5") {
			return nil, nil
		}

		if strings.Contains(line, "apiset:") {
			declaration.Side = strings.TrimRight(line[8:], " ")
		}

		if !isInCodeBlock && strings.HasPrefix(line, "```c") {
			isInCodeBlock = true
			continue
		}

		if isInCodeBlock && strings.HasPrefix(line, "//") {
			declaration.Hash = strings.Split(line, " ")[1]
			continue
		}

		if isInCodeBlock && strings.HasSuffix(line, ";") {
			regex := regexp.MustCompile(`\s*cs_type\([^)]*\)\s*`)
			line = regex.ReplaceAllString(line, "")
			line = strings.TrimRight(line, " ")

			split := strings.Split(line, " ")
			returnType := split[0]
			nativeDecl := strings.Join(split[1:], " ")

			strArgs := utils.SplitBetween(nativeDecl, "(", ");")
			if strArgs != "" {
				splitArgs := strings.Split(strArgs, ",")
				l := len(splitArgs)

				for i, _args := range splitArgs {
					_args = strings.TrimLeft(_args, " ")
					_definitions := strings.Split(_args, " ")

					if len(_definitions) < 2 {
						fmt.Println(fp)
					}

					_type := strings.Trim(_definitions[0], " ")
					_name := strings.Trim(_definitions[1], " ")
					if _name == "var" {
						_name = "_var"
					}

					if i < l-1 {
						declaration.TsArgs += fmt.Sprintf("%s: %s, ", _name, utils.ToTsArg(_type))
						declaration.Args += fmt.Sprintf("%s, ", _name)
					} else {
						declaration.TsArgs += fmt.Sprintf("%s: %s", _name, utils.ToTsArg(_type))
						declaration.Args += _name
					}
				}
			}

			name := strings.Replace(nativeDecl, fmt.Sprintf("(%s);", strArgs), "", 1)
			if name[0:1] == "_" && name[0:3] != "_0x" {
				name = name[1:]
			}

			declaration.Name = name
			declaration.ReturnValue = utils.ToTsArg(returnType)
			declaration.Call = fmt.Sprintf("%s(%s);", computeCfxNativeName(name), declaration.Args)

			switch declaration.ReturnValue {
			case "void":
				break
			case "Vector3":
				declaration.PreCall = "const result = "
				declaration.PostCall = "\n    return new Vector3(result[0], result[1], result[2]);"
				break
			default:
				if declaration.ReturnValue != "void" {
					declaration.PreCall = "return "
				}
				break
			}

			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return declaration, nil
}

func main() {
	gen, err := generator.InitGenerator()
	if err != nil {
		log.Fatal(err)
	}

	_fps1, err := utils.GetEveryFilesInFolder("deps/gta-natives")
	if err != nil {
		log.Fatal(err)
	}

	_fps2, err := utils.GetEveryFilesInFolder("deps/fivem-natives")
	if err != nil {
		log.Fatal(err)
	}

	gen.Fps = append(gen.Fps, _fps1...)
	gen.Fps = append(gen.Fps, _fps2...)

	for _, fp := range gen.Fps {
		declaration, err := parseGtaNative(fp)
		if err != nil {
			log.Fatal(err)
		}

		if declaration != nil {
			gen.Declarations = append(gen.Declarations, *declaration)
		}
	}

	header := "/* eslint-disable @typescript-eslint/no-explicit-any */\n\nimport { Vector3 } from \"shared\";\n"
	if _, err := fmt.Fprintln(gen.FileClient, header); err != nil {
		log.Fatal(err)
	}

	if _, err := fmt.Fprintln(gen.FileServer, header); err != nil {
		log.Fatal(err)
	}

	if err := gen.WriteDeclarations(); err != nil {
		log.Fatal(err)
	}
}
