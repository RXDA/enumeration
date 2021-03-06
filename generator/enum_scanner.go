package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/packagesx"

	"github.com/go-courier/enumeration"
)

func NewEnumScanner(pkg *packagesx.Package) *EnumScanner {
	return &EnumScanner{
		pkg: pkg,
	}
}

type EnumScanner struct {
	pkg     *packagesx.Package
	EnumSet map[*types.TypeName][]enumeration.EnumOption
}

func sortEnumOptions(enumOptions []enumeration.EnumOption) []enumeration.EnumOption {
	sort.Slice(enumOptions, func(i, j int) bool {
		return enumOptions[i].ConstValue > enumOptions[j].ConstValue
	})
	return enumOptions
}

func (scanner *EnumScanner) Enum(typeName *types.TypeName) []enumeration.EnumOption {
	if typeName == nil {
		return nil
	}

	if enumOptions, ok := scanner.EnumSet[typeName]; ok {
		return sortEnumOptions(enumOptions)
	}

	if !strings.Contains(typeName.Type().Underlying().String(), "int") {
		panic(fmt.Errorf("enumeration type underlying must be an int or uint, but got %s", typeName.String()))
	}

	pkgInfo := scanner.pkg.Pkg(typeName.Pkg().Path())
	if pkgInfo == nil {
		return nil
	}

	typeNameString := typeName.Name()

	for ident, def := range pkgInfo.TypesInfo.Defs {
		typeConst, ok := def.(*types.Const)
		if !ok {
			continue
		}
		if typeConst.Type() != typeName.Type() {
			continue
		}
		name := typeConst.Name()

		if strings.HasPrefix(name, "_") {
			continue
		}

		val := typeConst.Val()
		label := strings.TrimSpace(ident.Obj.Decl.(*ast.ValueSpec).Comment.Text())

		if strings.HasPrefix(name, codegen.UpperSnakeCase(typeNameString)) {
			var values = strings.SplitN(name, "__", 2)
			if len(values) == 2 {
				firstChar := values[1][0]
				if '0' <= firstChar && firstChar <= '9' {
					panic(fmt.Errorf("not support enum key starts with number, please fix `%s`", name))
				}
				intVal, _ := strconv.ParseInt(val.String(), 10, 64)
				scanner.addEnum(typeName, int(intVal), values[1], label)
			}
		}
	}

	return sortEnumOptions(scanner.EnumSet[typeName])
}

func (scanner *EnumScanner) addEnum(typeName *types.TypeName, constValue int, value string, label string) {
	if scanner.EnumSet == nil {
		scanner.EnumSet = map[*types.TypeName][]enumeration.EnumOption{}
	}

	if label == "" {
		label = value
	}

	scanner.EnumSet[typeName] = append(scanner.EnumSet[typeName], enumeration.EnumOption{
		ConstValue: constValue,
		Value:      value,
		Label:      label,
	})
}
