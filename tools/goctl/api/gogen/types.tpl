// Code generated by goctl. DO NOT EDIT.
// versions:
//  goctl version: {{.version}}

package types{{if .containsTime}}
import (
	"time"
){{end}}
{{.types}}
