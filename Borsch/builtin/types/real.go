package types

import (
	"errors"
	"math"
	"strconv"

	"github.com/YuriyLisovskiy/borsch-lang/Borsch/common"
	"github.com/YuriyLisovskiy/borsch-lang/Borsch/util"
)

type RealInstance struct {
	BuiltinInstance
	Value float64
}

func NewRealInstance(value float64) RealInstance {
	return RealInstance{
		BuiltinInstance: BuiltinInstance{
			ClassInstance: ClassInstance{
				class:      Real,
				attributes: map[string]common.Value{},
				address:    "",
			},
		},
		Value: value,
	}
}

func (t RealInstance) String(common.State) (string, error) {
	return strconv.FormatFloat(t.Value, 'f', -1, 64), nil
}

func (t RealInstance) Representation(state common.State) (string, error) {
	return t.String(state)
}

func (t RealInstance) AsBool(common.State) (bool, error) {
	return t.Value != 0.0, nil
}

func compareReals(_ common.State, op common.Operator, self, other common.Value) (int, error) {
	left, ok := self.(RealInstance)
	if !ok {
		return 0, util.IncorrectUseOfFunctionError("compareReals")
	}

	switch right := other.(type) {
	case NilInstance:
	case BoolInstance:
		rightVal := boolToFloat64(right.Value)
		if left.Value == rightVal {
			return 0, nil
		}

		if left.Value < rightVal {
			return -1, nil
		}

		return 1, nil
	case IntegerInstance:
		rightVal := float64(right.Value)
		if left.Value == rightVal {
			return 0, nil
		}

		if left.Value < rightVal {
			return -1, nil
		}

		return 1, nil
	case RealInstance:
		if left.Value == right.Value {
			return 0, nil
		}

		if left.Value < right.Value {
			return -1, nil
		}

		return 1, nil
	default:
		return 0, util.OperatorNotSupportedError(op, left.GetTypeName(), right.GetTypeName())
	}

	// -2 is something other than -1, 0 or 1 and means 'not equals'
	return -2, nil
}

func newRealBinaryOperator(
	name string,
	doc string,
	handler func(RealInstance, common.Value) (common.Value, error),
) *FunctionInstance {
	return newBinaryMethod(
		name,
		Real,
		Any,
		doc,
		func(_ common.State, left common.Value, right common.Value) (common.Value, error) {
			if leftInstance, ok := left.(RealInstance); ok {
				return handler(leftInstance, right)
			}

			return nil, util.IncorrectUseOfFunctionError(name)
		},
	)
}

func newRealUnaryOperator(
	name string,
	doc string,
	handler func(RealInstance) (common.Value, error),
) *FunctionInstance {
	return newUnaryMethod(
		name, Real, Any, doc, func(_ common.State, left common.Value) (common.Value, error) {
			if leftInstance, ok := left.(RealInstance); ok {
				return handler(leftInstance)
			}

			return nil, util.IncorrectUseOfFunctionError(name)
		},
	)
}

func newRealClass() *Class {
	initAttributes := func(attrs *map[string]common.Value) {
		*attrs = MergeAttributes(
			map[string]common.Value{
				// TODO: add doc
				common.ConstructorName: newBuiltinConstructor(Real, ToReal, ""),
				common.PowOp.Name(): newRealBinaryOperator(
					// TODO: add doc
					common.PowOp.Name(), "", func(self RealInstance, other common.Value) (common.Value, error) {
						switch o := other.(type) {
						case RealInstance:
							return NewRealInstance(math.Pow(self.Value, o.Value)), nil
						case IntegerInstance:
							return NewRealInstance(math.Pow(self.Value, float64(o.Value))), nil
						case BoolInstance:
							return NewRealInstance(math.Pow(self.Value, boolToFloat64(o.Value))), nil
						default:
							return nil, nil
						}
					},
				),
				common.UnaryPlus.Name(): newRealUnaryOperator(
					// TODO: add doc
					common.UnaryPlus.Name(), "", func(self RealInstance) (common.Value, error) {
						return self, nil
					},
				),
				common.UnaryMinus.Name(): newRealUnaryOperator(
					// TODO: add doc
					common.UnaryMinus.Name(), "", func(self RealInstance) (common.Value, error) {
						return NewRealInstance(-self.Value), nil
					},
				),
				common.MulOp.Name(): newRealBinaryOperator(
					// TODO: add doc
					common.MulOp.Name(), "", func(self RealInstance, other common.Value) (common.Value, error) {
						switch o := other.(type) {
						case BoolInstance:
							return NewRealInstance(self.Value * boolToFloat64(o.Value)), nil
						case IntegerInstance:
							return NewRealInstance(self.Value * float64(o.Value)), nil
						case RealInstance:
							return NewRealInstance(self.Value * o.Value), nil
						default:
							return nil, nil
						}
					},
				),
				common.DivOp.Name(): newRealBinaryOperator(
					// TODO: add doc
					common.DivOp.Name(), "", func(self RealInstance, other common.Value) (common.Value, error) {
						switch o := other.(type) {
						case BoolInstance:
							if o.Value {
								return NewRealInstance(self.Value), nil
							}
						case IntegerInstance:
							if o.Value != 0 {
								return NewRealInstance(self.Value / float64(o.Value)), nil
							}
						case RealInstance:
							if o.Value != 0.0 {
								return NewRealInstance(self.Value / o.Value), nil
							}
						default:
							return nil, nil
						}

						return nil, errors.New("ділення на нуль")
					},
				),
				common.AddOp.Name(): newRealBinaryOperator(
					// TODO: add doc
					common.AddOp.Name(), "", func(self RealInstance, other common.Value) (common.Value, error) {
						switch o := other.(type) {
						case BoolInstance:
							return NewRealInstance(self.Value + boolToFloat64(o.Value)), nil
						case IntegerInstance:
							return NewRealInstance(self.Value + float64(o.Value)), nil
						case RealInstance:
							return NewRealInstance(self.Value + o.Value), nil
						default:
							return nil, nil
						}
					},
				),
				common.SubOp.Name(): newRealBinaryOperator(
					// TODO: add doc
					common.SubOp.Name(), "", func(self RealInstance, other common.Value) (common.Value, error) {
						switch o := other.(type) {
						case BoolInstance:
							return NewRealInstance(self.Value - boolToFloat64(o.Value)), nil
						case IntegerInstance:
							return NewRealInstance(self.Value - float64(o.Value)), nil
						case RealInstance:
							return NewRealInstance(self.Value - o.Value), nil
						default:
							return nil, nil
						}
					},
				),
			},
			MakeLogicalOperators(Real),
			MakeComparisonOperators(Real, compareReals),
			MakeCommonOperators(Real),
		)
	}

	return &Class{
		Name:            common.RealTypeName,
		IsFinal:         true,
		Bases:           []*Class{},
		Parent:          BuiltinPackage,
		AttrInitializer: initAttributes,
		GetEmptyInstance: func() (common.Value, error) {
			return NewRealInstance(0), nil
		},
	}
}
