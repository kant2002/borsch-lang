package types

import "fmt"

type methodType int8

const (
	method   methodType = 0
	function methodType = 1
	lambda   methodType = 2
)

var (
	MethodClass   = ObjectClass.ClassNew("метод", map[string]Object{}, true, nil, nil)
	FunctionClass = ObjectClass.ClassNew("функція", map[string]Object{}, true, nil, nil)
	LambdaClass   = ObjectClass.ClassNew("лямбда", map[string]Object{}, true, nil, nil)
)

type MethodFunc func(ctx Context, args Tuple, kwargs StringDict) (Object, error)

type Method struct {
	Name    string
	Package *Package

	Parameters  []MethodParameter
	ReturnTypes []MethodReturnType

	methodF MethodFunc

	typ methodType
}

type MethodParameter struct {
	Class      *Class
	Classes    []*Class
	Name       string
	IsNullable bool
	IsVariadic bool
}

func (value *MethodParameter) accepts(class *Class) bool {
	if value.Class != nil && value.Class == class {
		return true
	}

	for _, cls := range value.Classes {
		if cls.Class() == class {
			return true
		}
	}

	return false
}

type MethodReturnType struct {
	Class      *Class
	IsNullable bool
}

func makeMethod(
	name string,
	pkg *Package,
	parameters []MethodParameter,
	returnTypes []MethodReturnType,
	methodF MethodFunc,
	typ methodType,
) *Method {
	if pkg == nil && Initialized {
		panic("package is nil")
	}

	if methodF == nil {
		panic("methodF is nil")
	}

	return &Method{
		Name:        name,
		Package:     pkg,
		Parameters:  parameters,
		ReturnTypes: returnTypes,
		methodF:     methodF,
		typ:         typ,
	}
}

func MethodNew(
	name string,
	pkg *Package,
	parameters []MethodParameter,
	returnTypes []MethodReturnType,
	methodF MethodFunc,
) *Method {
	return makeMethod(name, pkg, parameters, returnTypes, methodF, method)
}

func FunctionNew(
	name string,
	pkg *Package,
	parameters []MethodParameter,
	returnTypes []MethodReturnType,
	methodF MethodFunc,
) *Method {
	return makeMethod(name, pkg, parameters, returnTypes, methodF, function)
}

func LambdaNew(
	name string,
	pkg *Package,
	parameters []MethodParameter,
	returnTypes []MethodReturnType,
	methodF MethodFunc,
) *Method {
	return makeMethod(name, pkg, parameters, returnTypes, methodF, lambda)
}

func (value *Method) Class() *Class {
	switch value.typ {
	case method:
		return MethodClass
	case function:
		return FunctionClass
	case lambda:
		return LambdaClass
	default:
		panic("unreachable")
	}
}

func (value *Method) call(args Tuple) (Object, error) {
	pLen := len(value.Parameters)
	aLen := len(args)
	if pLen != aLen {
		return nil, NewErrorf("кількість параметрів не дорівнює кількості аргументів, %d != %d", pLen, aLen)
	}

	// TODO: take into account variable parameters!

	kwargs := StringDict{}
	for i, arg := range args {
		parameter := value.Parameters[i]
		if err := checkArg(&parameter, arg); err != nil {
			return nil, err
		}

		kwargs[parameter.Name] = arg
	}

	ctx := value.Package.Context.Derive()
	ctx.PushScope(kwargs)
	result, err := value.methodF(ctx, args, kwargs)
	if err != nil {
		return nil, err
	}

	// TODO: check result
	prefix := ""
	if len(value.ReturnTypes) > 1 {
		prefix = "одним із типів"
	} else {
		prefix = "типу"
	}

	names := ""
	for i, retType := range value.ReturnTypes {
		// Check for derived types and "any" type.
		if retType.Class == result.Class() {
			return result, nil
		}

		names += retType.Class.Name
		if i < len(value.ReturnTypes)-1 {
			names += ", "
		}
	}

	return nil, NewTypeErrorf("повернене значення має бути %s ʼ%sʼ, отримано ʼ%sʼ", prefix, names, result.Class().Name)
}

func (value *Method) IsMethod() bool {
	return value.typ == method
}

func (value *Method) IsFunction() bool {
	return value.typ == function
}

func (value *Method) IsLambda() bool {
	return value.typ == lambda
}

func checkArg(parameter *MethodParameter, arg Object) error {
	if parameter.accepts(AnyClass) {
		return nil
	}

	if arg == Nil {
		if parameter.accepts(NilClass) || parameter.IsNullable {
			return nil
		}

		return NewTypeErrorf("очікується ненульовий аргумент, отримано ʼ%sʼ", arg.Class().Name)
	}

	if parameter.accepts(arg.Class()) {
		return nil
	}

	appendix := ""
	if parameter.IsNullable {
		appendix = fmt.Sprintf("або ʼ%sʼ", NilClass.Name)
	}

	if parameter.Class != nil {
		return NewTypeErrorf(
			"очікується аргумент з типом ʼ%sʼ%s, отримано ʼ%sʼ",
			parameter.Class.Name,
			appendix,
			arg.Class().Name,
		)
	}

	var names string
	clsLen := len(parameter.Classes)
	for i, cls := range parameter.Classes {
		names += cls.Name
		if i < clsLen-1 {
			names += ", "
		}
	}

	return NewTypeErrorf(
		"очікується аргумент з одним із типів ʼ%sʼ%s, отримано ʼ%sʼ",
		names,
		appendix,
		arg.Class().Name,
	)
}

func checkReturnValue(cls *Class, returnValue Object) error {
	// TODO:
	return nil
}
