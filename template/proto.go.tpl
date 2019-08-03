syntax = "proto3";
package {{.Models}};

// The {{.Models}} service definition.
service {{.Name}} {
 {{range .Funcs }} rpc {{.Name}}({{.RequestName}}) returns () { {{.ResponseName}} }
{{ end }}
}
{{range .MessageList }}
message {{.Name}} {
{{range .MessageDetail }} {{.TypeName}} {{.AttrName}}={{.Num}}
{{ end }}
}
{{ end }}

