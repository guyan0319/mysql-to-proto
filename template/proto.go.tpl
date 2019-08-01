syntax = "proto3";
package {{.Models}};

// The {{.Models}} service definition.
service {{.Name}} {
 // {{range .Funcs }}
   //     rpc {{.Name}}({{.RequestName}}) returns () {{{.ResponseName}}}
  //{{ end }}
}

message UserRequest {
  uint64 id = 1;
  string name = 2;
  string email = 3;
  string phone= 4;
}

message UserResponse {
  uint64 id = 1;
  bool success = 2;
}
message UserFilter {
  uint64 id = 1;
}

