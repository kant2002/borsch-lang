package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/YuriyLisovskiy/borsch-lang/Borsch/common"
	"github.com/YuriyLisovskiy/borsch-lang/Borsch/util"
)

func getIndex(index, length int64) (int64, error) {
	if index >= 0 && index < length {
		return index, nil
	} else if index < 0 && index >= -length {
		return length + index, nil
	}

	return 0, errors.New("індекс за межами послідовності")
}

func normalizeBound(bound, length int64) int64 {
	if bound < 0 {
		return length + bound
	}

	return bound
}

func CheckResult(state common.State, result common.Type, function *FunctionInstance) error {
	if len(function.ReturnTypes) == 1 {
		err := checkSingleResult(state, result, function.ReturnTypes[0], function.Name)
		if err != nil {
			return errors.New(fmt.Sprintf(err.Error(), ""))
		}

		return nil
	}

	switch value := result.(type) {
	case ListInstance:
		if int64(len(function.ReturnTypes)) != value.Length(state) {
			var expectedTypes []string
			for _, retType := range function.ReturnTypes {
				expectedTypes = append(expectedTypes, retType.String())
			}

			var typesGot []string
			for _, retType := range value.Values {
				typesGot = append(typesGot, retType.GetTypeName())
			}

			return util.RuntimeError(
				fmt.Sprintf(
					"'%s' повертає значення з типами (%s), отримано (%s)",
					function.Name,
					strings.Join(expectedTypes, ", "),
					strings.Join(typesGot, ", "),
				),
			)
		}

		// TODO: check values in list
		for i, returnType := range function.ReturnTypes {
			if err := checkSingleResult(state, value.Values[i], returnType, function.Name); err != nil {
				return errors.New(fmt.Sprintf(err.Error(), fmt.Sprintf(" на позиції %d", i+1)))
			}
		}
	default:
		var expectedTypes []string
		for _, retType := range function.ReturnTypes {
			expectedTypes = append(expectedTypes, retType.String())
		}

		return util.RuntimeError(
			fmt.Sprintf(
				"'%s()' повертає значення з типами '(%s)', отримано '%s'",
				function.Name,
				strings.Join(expectedTypes, ", "),
				value.GetTypeName(),
			),
		)
	}

	return nil
}

func makeFuncSignature(funcName string) string {
	if funcName == "" {
		return "лямбда-вираз"
	}

	return fmt.Sprintf("'%s()", funcName)
}

func checkSingleResult(
	state common.State,
	result common.Type,
	returnType FunctionReturnType,
	funcName string,
) error {
	if result.(ObjectInstance).GetPrototype() == Nil {
		if returnType.Type != Nil && !returnType.IsNullable {
			resultStr, err := result.String(state)
			if err != nil {
				return err
			}

			return util.RuntimeError(
				fmt.Sprintf(
					"%s повертає ненульове значення%s, отримано '%s'",
					makeFuncSignature(funcName),
					"%s",
					resultStr,
				),
			)
		}
	} else if returnType.Type != Any && result.(ObjectInstance).GetPrototype() != returnType.Type {
		return util.RuntimeError(
			fmt.Sprintf(
				"%s повертає значення типу '%s'%s, отримано значення з типом '%s'",
				makeFuncSignature(funcName), returnType.String(), "%s", result.GetTypeName(),
			),
		)
	}

	return nil
}

func CheckFunctionArguments(
	state common.State,
	function *FunctionInstance,
	args *[]common.Type,
	_ *map[string]common.Type,
) error {
	parametersLen := len(*args)
	argsLen := len(function.Parameters)
	if argsLen > 0 && function.Parameters[argsLen-1].IsVariadic {
		argsLen--
		if parametersLen > argsLen {
			parametersLen = argsLen
		}
	}

	if parametersLen != argsLen {
		diffLen := argsLen - parametersLen
		if diffLen > 0 {
			end1 := "ій"
			end2 := "ий"
			end3 := ""
			if diffLen > 1 && diffLen < 5 {
				end1 = "і"
				end2 = "і"
				end3 = "и"
			} else if diffLen != 1 {
				end1 = "і"
				end2 = "их"
				end3 = "ів"
			}

			parametersStr := ""
			for c := parametersLen; c < argsLen; c++ {
				parametersStr += fmt.Sprintf("'%s'", function.Parameters[c].Name)
				if c < argsLen-2 {
					parametersStr += ", "
				} else if c < argsLen-1 {
					parametersStr += " та "
				}
			}

			return util.RuntimeError(
				fmt.Sprintf(
					"при виклику '%s()' відсутн%s %d необхідн%s параметр%s: %s",
					function.Name, end1, diffLen, end2, end3, parametersStr,
				),
			)
		} else {
			end1 := "ий"
			end2 := ""
			if argsLen > 1 && argsLen < 5 {
				end1 = "і"
				end2 = "и"
			} else if argsLen != 1 {
				end1 = "их"
				end2 = "ів"
			}

			return util.RuntimeError(
				fmt.Sprintf(
					"'%s()' приймає %d необхідн%s параметр%s, отримано %d",
					function.Name, argsLen, end1, end2, parametersLen,
				),
			)
		}
	}

	var c int
	for c = 0; c < argsLen; c++ {
		parameter := function.Parameters[c]
		if parameter.Type == Any {
			continue
		}

		arg := (*args)[c]
		argPrototype := arg.(ObjectInstance).GetPrototype()
		if argPrototype == Nil && parameter.IsNullable {
			continue
		}

		if parameter.Type == argPrototype {
			continue
		}

		return util.RuntimeError(
			fmt.Sprintf(
				"аргумент '%s' очікує параметр з типом '%s', отримано '%s'",
				parameter.Name, parameter.GetTypeName(), arg.GetTypeName(),
			),
		)

		// if argPrototype == Nil {
		// 	if function.Parameters[c].Type != Nil && !function.Parameters[c].IsNullable {
		// 		argStr, err := arg.String(state)
		// 		if err != nil {
		// 			return err
		// 		}
		//
		// 		return util.RuntimeError(
		// 			fmt.Sprintf(
		// 				"аргумент '%s' очікує ненульовий параметр, отримано '%s'",
		// 				function.Parameters[c].Name,
		// 				argStr,
		// 			),
		// 		)
		// 	}
		// } else if function.Parameters[c].Type != Any && argPrototype != function.Parameters[c].Type {
		// 	return util.RuntimeError(
		// 		fmt.Sprintf(
		// 			"аргумент '%s' очікує параметр з типом '%s', отримано '%s'",
		// 			function.Parameters[c].Name, function.Parameters[c].GetTypeName(), arg.GetTypeName(),
		// 		),
		// 	)
		// }
	}

	if len(function.Parameters) > 0 {
		if lastArgument := function.Parameters[len(function.Parameters)-1]; lastArgument.IsVariadic {
			if len(*args)-parametersLen > 0 {
				parametersLen = len(*args)
				for k := c; k < parametersLen; k++ {
					arg := (*args)[k]
					argPrototype := arg.(ObjectInstance).GetPrototype()
					if argPrototype == Nil {
						if lastArgument.Type != Nil && !lastArgument.IsNullable {
							argStr, err := arg.String(state)
							if err != nil {
								return err
							}

							return util.RuntimeError(
								fmt.Sprintf(
									"аргумент '%s' очікує ненульовий параметр, отримано '%s'",
									lastArgument.Name,
									argStr,
								),
							)
						}
					} else if lastArgument.Type != nil && argPrototype != lastArgument.Type {
						return util.RuntimeError(
							fmt.Sprintf(
								"аргумент '%s' очікує список параметрів з типом '%s', отримано '%s'",
								lastArgument.Name,
								lastArgument.GetTypeName(),
								argPrototype.GetTypeName(),
							),
						)
					}
				}
			}
		}
	}

	return nil
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}

	return 0
}

func boolToFloat64(v bool) float64 {
	if v {
		return 1.0
	}

	return 0.0
}

func getAttributes(attributes map[string]common.Type) (DictionaryInstance, error) {
	dict := NewDictionaryInstance()
	for key, val := range attributes {
		err := dict.SetElement(NewStringInstance(key), val)
		if err != nil {
			return DictionaryInstance{}, err
		}
	}

	return dict, nil
}

func getLength(state common.State, sequence common.Type) (int64, error) {
	switch self := sequence.(type) {
	case common.SequentialType:
		return self.Length(state), nil
	}

	return 0, errors.New(fmt.Sprint("invalid type in length operator: ", sequence.GetTypeName()))
}

func mergeAttributes(a map[string]common.Type, b ...map[string]common.Type) map[string]common.Type {
	for _, m := range b {
		for key, val := range m {
			a[key] = val
		}
	}

	return a
}

func newBinaryMethod(
	name string,
	selfType *Class,
	returnType *Class,
	doc string,
	handler func(common.State, common.Type, common.Type) (common.Type, error),
) *FunctionInstance {
	return NewFunctionInstance(
		name,
		[]FunctionParameter{
			{
				Type:       selfType,
				Name:       "я",
				IsVariadic: false,
				IsNullable: false,
			},
			{
				Type:       Any,
				Name:       "інший",
				IsVariadic: false,
				IsNullable: true,
			},
		},
		func(state common.State, args *[]common.Type, _ *map[string]common.Type) (common.Type, error) {
			return handler(state, (*args)[0], (*args)[1])
		},
		[]FunctionReturnType{
			{
				Type:       returnType,
				IsNullable: false,
			},
		},
		true,
		nil,
		doc,
	)
}

func newUnaryMethod(
	name string,
	selfType *Class,
	returnType *Class,
	doc string,
	handler func(common.State, common.Type) (common.Type, error),
) *FunctionInstance {
	return NewFunctionInstance(
		name,
		[]FunctionParameter{
			{
				Type:       selfType,
				Name:       "я",
				IsVariadic: false,
				IsNullable: false,
			},
		},
		func(state common.State, args *[]common.Type, _ *map[string]common.Type) (common.Type, error) {
			return handler(state, (*args)[0])
		},
		[]FunctionReturnType{
			{
				Type:       returnType,
				IsNullable: false,
			},
		},
		true,
		nil,
		doc,
	)
}

func newComparisonOperator(
	operator common.Operator,
	itemType *Class,
	doc string,
	comparator func(common.State, common.Type, common.Type) (int, error),
	checker func(res int) bool,
) *FunctionInstance {
	return newBinaryMethod(
		operator.Name(),
		itemType,
		Bool,
		doc,
		func(state common.State, self common.Type, other common.Type) (common.Type, error) {
			res, err := comparator(state, self, other)
			if err != nil {
				return nil, err
			}

			return NewBoolInstance(checker(res)), nil
		},
	)
}

func makeComparisonOperators(
	itemType *Class,
	comparator func(common.State, common.Type, common.Type) (int, error),
) map[string]common.Type {
	return map[string]common.Type{
		common.EqualsOp.Name(): newComparisonOperator(
			// TODO: add doc
			common.EqualsOp, itemType, "", comparator, func(res int) bool {
				return res == 0
			},
		),
		common.NotEqualsOp.Name(): newComparisonOperator(
			// TODO: add doc
			common.NotEqualsOp, itemType, "", comparator, func(res int) bool {
				return res != 0
			},
		),
		common.GreaterOp.Name(): newComparisonOperator(
			// TODO: add doc
			common.GreaterOp, itemType, "", comparator, func(res int) bool {
				return res == 1
			},
		),
		common.GreaterOrEqualsOp.Name(): newComparisonOperator(
			// TODO: add doc
			common.GreaterOrEqualsOp, itemType, "", comparator, func(res int) bool {
				return res == 0 || res == 1
			},
		),
		common.LessOp.Name(): newComparisonOperator(
			// TODO: add doc
			common.LessOp, itemType, "", comparator, func(res int) bool {
				return res == -1
			},
		),
		common.LessOrEqualsOp.Name(): newComparisonOperator(
			// TODO: add doc
			common.LessOrEqualsOp, itemType, "", comparator, func(res int) bool {
				return res == 0 || res == -1
			},
		),
	}
}

func makeLogicalOperators(itemType *Class) map[string]common.Type {
	return map[string]common.Type{
		common.NotOp.Name(): newUnaryMethod(
			// TODO: add doc
			common.NotOp.Name(),
			itemType,
			Bool,
			"",
			func(state common.State, self common.Type) (common.Type, error) {
				selfBool, err := self.AsBool(state)
				if err != nil {
					return nil, err
				}

				return NewBoolInstance(!selfBool), nil
			},
		),
		common.AndOp.Name(): newBinaryMethod(
			// TODO: add doc
			common.AndOp.Name(),
			itemType,
			Bool,
			"",
			func(state common.State, self common.Type, other common.Type) (common.Type, error) {
				selfBool, err := self.AsBool(state)
				if err != nil {
					return nil, err
				}

				otherBool, err := other.AsBool(state)
				if err != nil {
					return nil, err
				}

				return NewBoolInstance(selfBool && otherBool), nil
			},
		),
		common.OrOp.Name(): newBinaryMethod(
			// TODO: add doc
			common.OrOp.Name(),
			itemType,
			Bool,
			"",
			func(state common.State, self common.Type, other common.Type) (common.Type, error) {
				selfBool, err := self.AsBool(state)
				if err != nil {
					return nil, err
				}

				otherBool, err := other.AsBool(state)
				if err != nil {
					return nil, err
				}

				return NewBoolInstance(selfBool || otherBool), nil
			},
		),
	}
}

func makeCommonOperators(itemType *Class) map[string]common.Type {
	return map[string]common.Type{
		// TODO: add doc
		common.BoolOperatorName: newUnaryMethod(
			common.BoolOperatorName, itemType, Bool, "",
			func(state common.State, self common.Type) (common.Type, error) {
				boolVal, err := self.AsBool(state)
				if err != nil {
					return nil, err
				}

				return NewBoolInstance(boolVal), nil
			},
		),
	}
}

func newBuiltinConstructor(
	itemType *Class,
	handler func(common.State, ...common.Type) (common.Type, error),
	doc string,
) *FunctionInstance {
	return NewFunctionInstance(
		common.ConstructorName,
		[]FunctionParameter{
			{
				Type:       itemType,
				Name:       "я",
				IsVariadic: false,
				IsNullable: false,
			},
			{
				Type:       Any,
				Name:       "значення",
				IsVariadic: true,
				IsNullable: true,
			},
		},
		func(state common.State, args *[]common.Type, _ *map[string]common.Type) (common.Type, error) {
			self, err := handler(state, (*args)[1:]...)
			if err != nil {
				return nil, err
			}

			(*args)[0] = self
			return NewNilInstance(), nil
		},
		[]FunctionReturnType{
			{
				Type:       Nil,
				IsNullable: false,
			},
		},
		true,
		nil,
		doc,
	)
}

func newLengthOperator(
	itemType *Class,
	handler func(common.State, common.Type) (int64, error),
	doc string,
) *FunctionInstance {
	return NewFunctionInstance(
		common.LengthOperatorName,
		[]FunctionParameter{
			{
				Type:       itemType,
				Name:       "я",
				IsVariadic: false,
				IsNullable: false,
			},
		},
		func(state common.State, args *[]common.Type, _ *map[string]common.Type) (common.Type, error) {
			length, err := handler(state, (*args)[0])
			if err != nil {
				return nil, err
			}

			return NewIntegerInstance(length), nil
		},
		[]FunctionReturnType{
			{
				Type:       Integer,
				IsNullable: false,
			},
		},
		true,
		nil,
		doc,
	)
}
