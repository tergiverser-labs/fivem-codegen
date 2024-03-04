package utils

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Joaat(s string) uint32 {
	k := strings.ToLower(s)
	var h uint32

	for _, char := range k {
		h += uint32(char)
		h += h << 10
		h ^= h >> 6
	}

	h += h << 3
	h ^= h >> 11
	h += h << 15

	return h
}

func SnakeToCamelCase(input string) string {
	s := input[0:1] == "_"
	if s {
		input = input[1:]
	}

	words := strings.Split(strings.ToLower(input), "_")
	for i := 1; i < len(words); i++ {
		words[i] = strings.ToUpper(string(words[i][0])) + words[i][1:]
	}

	join := strings.Join(words, "")
	if s {
		return "_" + join
	}

	return join
}

func FirstToUpper(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}

	lc := unicode.ToUpper(r)
	if r == lc {
		return s
	}

	return string(lc) + s[size:]
}

func SplitBetween(str, bef, aft string) string {
	sa := strings.SplitN(str, bef, 2)
	if len(sa) == 1 {
		return ""
	}
	sa = strings.SplitN(sa[1], aft, 2)
	if len(sa) == 1 {
		return ""
	}
	return sa[0]
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(e, a) {
			return true
		}
	}

	return false
}

func GetEveryFilesInFolder(root string) ([]string, error) {
	fps := make([]string, 0)
	walk := func(fp string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && !Contains([]string{".git", ".ci", ".github", ".gitignore", "README.md"}, fp) {
			fps = append(fps, fp)
		}

		return nil
	}

	if err := filepath.WalkDir(root, walk); err != nil {
		return nil, err
	}

	return fps, nil
}

func ToTsArg(arg string) string {
	switch arg {
	case "int", "float", "Blip", "Cam", "Entity", "FireId", "Interior", "ItemSet", "Object", "Ped", "Pickup", "Player", "Vehicle", "ScrHandle":
		return "number"

	case "int*", "float*", "Blip*", "Cam*", "Entity*", "FireId*", "Interior*", "ItemSet*", "Object*", "Ped*", "Pickup*", "Player*", "Vehicle*", "ScrHandle*":
		return "number | null"

	case "intPtr", "floatPtr", "BlipPtr", "CamPtr", "EntityPtr", "FireIdPtr", "InteriorPtr", "ItemSetPtr", "ObjectPtr", "PedPtr", "PickupPtr", "PlayerPtr", "VehiclePtr", "ScrHandlePtr":
		return "number | null"

	case "Hash":
		return "number | string"

	case "Hash*", "HashPtr":
		return "number | string | null"

	case "BOOL":
		return "boolean"

	case "BOOL*", "BOOLPtr", "bool*":
		return "boolean | null"

	case "char*", "const char*", "char", "charPtr":
		return "string"

	case "Any", "Any*", "AnyPtr":
		return "any"

	case "Vector3":
		return "Vector3"

	case "Vector3*":
		return "Vector3 | null"

	case "void":
		return "void"

	// fivem custom natives
	case "func":
		return "(...args: any[]) => any"

	case "long":
		return "number"

	case "bool":
		return "boolean"

	case "object":
		return "Record<string, unknown>"
	}

	fmt.Printf("unsupported arg type %s\n", arg)
	return ""
}
