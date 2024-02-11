package builddefinition

type funcsContext struct {
	args map[string]string
}

func funcs(ctx funcsContext) []nativeFunc {
	return []nativeFunc{
		nativeFunc2[string, string]{
			name: "cache",
			v1: argDesc{
				name: "id",
			},
			v2: argDesc{
				name: "path",
			},
			f: func(id string, path string) (interface{}, error) {
				return map[string]interface{}{
					"id":   id,
					"path": path,
				}, nil
			},
		},
		nativeFunc2[string, string]{
			name: "copy",
			v1: argDesc{
				name: "src",
			},
			v2: argDesc{
				name: "dst",
			},
			f: func(src string, dst string) (interface{}, error) {
				return map[string]interface{}{
					"src": src,
					"dst": dst,
				}, nil
			},
		},
		nativeFunc3[string, string, string]{
			name: "copyFrom",
			v1: argDesc{
				name: "from",
			},
			v2: argDesc{
				name: "src",
			},
			v3: argDesc{
				name: "dst",
			},
			f: func(from string, src string, dst string) (interface{}, error) {
				return map[string]interface{}{
					"from": from,
					"src":  src,
					"dst":  dst,
				}, nil
			},
		},
		nativeFunc2[string, string]{
			name: "secret",
			v1: argDesc{
				name: "id",
			},
			v2: argDesc{
				name: "path",
			},
			f: func(id string, path string) (interface{}, error) {
				return map[string]interface{}{
					"id":   id,
					"path": path,
				}, nil
			},
		},
		nativeFunc1[string]{
			name: "defArgSet",
			v: argDesc{
				name: "defArg",
			},
			f: func(v string) (interface{}, error) {
				_, exists := ctx.args[v]
				return exists, nil
			},
		},
		nativeFunc1[string]{
			name: "defArg",
			v: argDesc{
				name: "defArg",
			},
			f: func(v string) (interface{}, error) {
				arg := ctx.args[v]
				return arg, nil
			},
		},
	}
}
