package block

import (
	"encoding/json"
	"fmt"
)

// Reply provides a standardized way to respond to CLI commands, supporting both
// plain text and JSON output formats.
func Reply(isJSON bool, status, exe string) {
	if isJSON {
		out, _ := json.Marshal(map[string]string{"status": status, "exe": exe})
		fmt.Println(string(out))
	} else {
		fmt.Println(status+":", exe)
	}
}

// ReplyList provides a standardized way to output a list of strings, supporting
// both plain text and JSON formats.
func ReplyList(isJSON bool, list []string) {
	if isJSON {
		out, _ := json.Marshal(list)
		fmt.Println(string(out))
	} else {
		if len(list) == 0 {
			fmt.Println("block-list is empty")
			return
		}
		fmt.Println("blocked programs:")
		for _, v := range list {
			fmt.Println(" -", v)
		}
	}
}
