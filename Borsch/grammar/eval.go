package grammar

import (
	"errors"
	"fmt"

	"github.com/YuriyLisovskiy/borsch-lang/Borsch/common"
	"github.com/YuriyLisovskiy/borsch-lang/Borsch/ops"
	"github.com/YuriyLisovskiy/borsch-lang/Borsch/types"
	"github.com/YuriyLisovskiy/borsch-lang/Borsch/util"
	"github.com/alecthomas/participle/v2/lexer"
)

type Scope map[string]common.Type

func (p *Package) Evaluate(ctx common.Context) (common.Type, error) {
	ctx.PushScope(Scope{})
	for _, stmt := range p.Stmts {
		_, _, err := stmt.Evaluate(ctx, false)
		if err != nil {
			pos := stmt.getPos()
			return nil, errors.New(
				fmt.Sprintf(
					"  Файл \"%s\", рядок %d, позиція %d,\n    %s\n%s",
					pos.Filename, pos.Line, pos.Column, stmt.String(), err.Error(),
				),
			)
		}
	}

	if err := ctx.BuildPackage(); err != nil {
		return nil, err
	}

	ctx.PopScope()
	return ctx.GetPackage(), nil
}

func (s *WhileStmt) Evaluate(ctx common.Context) (common.Type, bool, error) {
	// TODO:
	panic("unreachable")
}

func (s *IfStmt) Evaluate(ctx common.Context, inFunction bool) (common.Type, bool, error) {
	if s.Condition != nil {
		condition, err := s.Condition.Evaluate(ctx, nil)
		if err != nil {
			return nil, false, err
		}

		if condition.AsBool() {
			ctx.PushScope(Scope{})
			result, forceReturn, err := s.Body.Evaluate(ctx, inFunction)
			if err != nil {
				return nil, false, err
			}

			ctx.PopScope()
			return result, forceReturn, nil
		}

		if len(s.ElseIfStmts) != 0 {
			gotResult := false
			var result common.Type = nil
			var err error = nil
			for _, stmt := range s.ElseIfStmts {
				ctx.PushScope(Scope{})
				var forceReturn bool
				gotResult, result, forceReturn, err = stmt.Evaluate(ctx, inFunction)
				if err != nil {
					return nil, false, err
				}

				ctx.PopScope()
				if forceReturn {
					return result, true, nil
				}

				if gotResult {
					break
				}
			}

			if gotResult {
				return result, false, nil
			}
		}

		if s.Else != nil {
			ctx.PushScope(Scope{})
			result, forceReturn, err := s.Else.Evaluate(ctx, inFunction)
			if err != nil {
				return nil, false, err
			}

			ctx.PopScope()
			return result, forceReturn, nil
		}

		return nil, false, nil
	}

	return nil, false, errors.New("interpreter: condition is nil")
}

func (s *ElseIfStmt) Evaluate(ctx common.Context, inFunction bool) (bool, common.Type, bool, error) {
	condition, err := s.Condition.Evaluate(ctx, nil)
	if err != nil {
		return false, nil, false, err
	}

	if condition.AsBool() {
		ctx.PushScope(Scope{})
		result, forceReturn, err := s.Body.Evaluate(ctx, inFunction)
		if err != nil {
			return false, nil, false, err
		}

		ctx.PopScope()
		return true, result, forceReturn, nil
	}

	return false, nil, false, nil
}

// Evaluate executes block of statements.
// Returns (result value, force stop flag, error)
func (b *BlockStmts) Evaluate(ctx common.Context, inFunction bool) (common.Type, bool, error) {
	for _, stmt := range b.Stmts {
		result, forceReturn, err := stmt.Evaluate(ctx, inFunction)
		if err != nil {
			return nil, false, err
		}

		if forceReturn || stmt.ReturnStmt != nil {
			return result, true, nil
		}
	}

	return types.NewNilInstance(), false, nil
}

// Evaluate executes statement.
// Returns (result value, force stop flag, error)
func (s *Stmt) Evaluate(ctx common.Context, inFunction bool) (common.Type, bool, error) {
	if s.IfStmt != nil {
		return s.IfStmt.Evaluate(ctx, inFunction)
	} else if s.WhileStmt != nil {
		return s.WhileStmt.Evaluate(ctx)
	} else if s.Block != nil {
		ctx.PushScope(Scope{})
		result, forceReturn, err := s.Block.Evaluate(ctx, inFunction)
		if err != nil {
			return nil, false, err
		}

		ctx.PopScope()
		return result, forceReturn, nil
	} else if s.FunctionDef != nil {
		function, err := s.FunctionDef.Evaluate(ctx)
		if err != nil {
			return nil, false, err
		}

		return function, false, ctx.SetVar(s.FunctionDef.Name, function)
	} else if s.ReturnStmt != nil {
		if !inFunction {
			return nil, false, errors.New("'повернути' за межами функції")
		}

		result, err := s.ReturnStmt.Evaluate(ctx)
		return result, false, err
		// } else if s.Expression != nil {
		//	result, err := s.Expression.Evaluate(ctx, nil)
		//	return result, false, err
	} else if s.Assignment != nil {
		result, err := s.Assignment.Evaluate(ctx)
		return result, false, err
	} else if s.Empty {
		return nil, false, nil
	}

	panic("unreachable")
}

func (s *Stmt) getPos() lexer.Position {
	if s.IfStmt != nil {
		return s.IfStmt.Pos
	} else if s.WhileStmt != nil {
		return s.WhileStmt.Pos
	} else if s.Block != nil {
		return s.Block.Pos
	} else if s.FunctionDef != nil {
		return s.FunctionDef.Pos
	} else if s.ReturnStmt != nil {
		return s.ReturnStmt.Pos
	} else if s.Assignment != nil {
		return s.Assignment.Pos
	} else if s.Empty {
		return s.Pos
	}

	panic("unreachable")
}

func (s *Stmt) String() string {
	if s.IfStmt != nil {
		return "s.IfStmt."
	} else if s.WhileStmt != nil {
		return "s.WhileStmt."
	} else if s.Block != nil {
		return "s.Block."
	} else if s.FunctionDef != nil {
		return "s.FunctionDef."
	} else if s.ReturnStmt != nil {
		return "повернути ..."
	} else if s.Assignment != nil {
		return "s.Assignment."
	} else if s.Empty {
		return ";"
	}

	panic("unreachable")
}

func (b *FunctionBody) Evaluate(ctx common.Context) (common.Type, error) {
	result, _, err := b.Stmts.Evaluate(ctx, true)
	return result, err
}

func (f *FunctionDef) Evaluate(ctx common.Context) (common.Type, error) {
	arguments := evalParameters(ctx, f.Parameters)
	returnTypes := evalReturnTypes(ctx, f.ReturnTypes)
	return types.NewFunctionInstance(
		f.Name,
		arguments,
		func(context interface{}, _ *[]common.Type, kwargs *map[string]common.Type) (common.Type, error) {
			funcContext := context.(common.FunctionContext)
			return f.Body.Evaluate(funcContext.Context)
		},
		returnTypes,
		false,
		ctx.GetPackage().(*types.PackageInstance),
		"", // TODO: add doc
	), nil
}

func (p *Parameter) Evaluate(_ common.Context) types.FunctionArgument {
	return types.FunctionArgument{
		TypeHash:   types.GetTypeHash(p.Type), // TODO: get type hash with package name
		Name:       p.Name,
		IsVariadic: false,
		IsNullable: p.IsNullable,
	}
}

func (t *ReturnType) Evaluate(ctx common.Context) types.FunctionReturnType {
	return types.FunctionReturnType{
		TypeHash:   types.GetTypeHash(t.Name), // TODO: get type hash with package name
		IsNullable: t.IsNullable,
	}
}

func (s *ReturnStmt) Evaluate(ctx common.Context) (common.Type, error) {
	resultCount := len(s.Expressions)
	switch {
	case resultCount == 1:
		return s.Expressions[0].Evaluate(ctx, nil)
	case resultCount > 1:
		result := types.NewListInstance()
		for _, expression := range s.Expressions {
			value, err := expression.Evaluate(ctx, nil)
			if err != nil {
				return nil, err
			}

			result.Values = append(result.Values, value)
		}

		return result, nil
	}

	panic("unreachable")
}

func (e *Expression) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	if e.LogicalAnd != nil {
		return e.LogicalAnd.Evaluate(ctx, valueToSet)
	}

	panic("unreachable")
}

func (a *Assignment) Evaluate(ctx common.Context) (common.Type, error) {
	if len(a.Next) == 0 {
		return a.Expression[0].Evaluate(ctx, nil)
	}

	return unpack(ctx, a.Expression, a.Next)
}

// Evaluate executes LogicalAnd operation.
// If `valueToSet` is nil, return variable or value from context,
// set a new value or return an error otherwise.
func (a *LogicalAnd) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	return evalBinaryOperator(ctx, valueToSet, ops.AndOp.Caption(), a.LogicalOr, a.Next)
}

func (a *LogicalOr) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	return evalBinaryOperator(ctx, valueToSet, ops.OrOp.Caption(), a.LogicalNot, a.Next)
}

func (a *LogicalNot) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	if a.Comparison != nil {
		return a.Comparison.Evaluate(ctx, valueToSet)
	}

	if a.Next != nil {
		value, err := a.Next.Evaluate(ctx, nil)
		if err != nil {
			return nil, err
		}

		return callMethod(value, ops.NotOp.Caption(), &[]common.Type{}, nil)
	}

	panic("unreachable")
}

func (a *Comparison) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	switch a.Op {
	case ">=":
		return evalBinaryOperator(ctx, valueToSet, ops.GreaterOrEqualsOp.Caption(), a.BitwiseOr, a.Next)
	case ">":
		return evalBinaryOperator(ctx, valueToSet, ops.GreaterOp.Caption(), a.BitwiseOr, a.Next)
	case "<=":
		return evalBinaryOperator(ctx, valueToSet, ops.LessOrEqualsOp.Caption(), a.BitwiseOr, a.Next)
	case "<":
		return evalBinaryOperator(ctx, valueToSet, ops.LessOp.Caption(), a.BitwiseOr, a.Next)
	case "==":
		return evalBinaryOperator(ctx, valueToSet, ops.EqualsOp.Caption(), a.BitwiseOr, a.Next)
	case "!=":
		return evalBinaryOperator(ctx, valueToSet, ops.NotEqualsOp.Caption(), a.BitwiseOr, a.Next)
	default:
		return a.BitwiseOr.Evaluate(ctx, valueToSet)
	}
}

func (a *BitwiseOr) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	return evalBinaryOperator(ctx, valueToSet, ops.BitwiseOrOp.Caption(), a.BitwiseXor, a.Next)
}

func (a *BitwiseXor) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	return evalBinaryOperator(ctx, valueToSet, ops.BitwiseXorOp.Caption(), a.BitwiseAnd, a.Next)
}

func (a *BitwiseAnd) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	return evalBinaryOperator(ctx, valueToSet, ops.BitwiseAndOp.Caption(), a.BitwiseShift, a.Next)
}

func (a *BitwiseShift) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	switch a.Op {
	case "<<":
		return evalBinaryOperator(ctx, valueToSet, ops.BitwiseLeftShiftOp.Caption(), a.Addition, a.Next)
	case ">>":
		return evalBinaryOperator(ctx, valueToSet, ops.BitwiseRightShiftOp.Caption(), a.Addition, a.Next)
	default:
		return a.Addition.Evaluate(ctx, valueToSet)
	}
}

func (a *Addition) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	switch a.Op {
	case "+":
		return evalBinaryOperator(ctx, valueToSet, ops.AddOp.Caption(), a.MultiplicationOrMod, a.Next)
	case "-":
		return evalBinaryOperator(ctx, valueToSet, ops.SubOp.Caption(), a.MultiplicationOrMod, a.Next)
	default:
		return a.MultiplicationOrMod.Evaluate(ctx, valueToSet)
	}
}

func (a *MultiplicationOrMod) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	switch a.Op {
	case "/":
		return evalBinaryOperator(ctx, valueToSet, ops.DivOp.Caption(), a.Unary, a.Next)
	case "*":
		return evalBinaryOperator(ctx, valueToSet, ops.MulOp.Caption(), a.Unary, a.Next)
	case "%":
		return evalBinaryOperator(ctx, valueToSet, ops.ModuloOp.Caption(), a.Unary, a.Next)
	default:
		return a.Unary.Evaluate(ctx, valueToSet)
	}
}

func (a *Unary) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	switch a.Op {
	case "+":
		return evalUnaryOperator(ctx, ops.UnaryPlus.Caption(), a.Next)
	case "-":
		return evalUnaryOperator(ctx, ops.UnaryMinus.Caption(), a.Next)
	case "~":
		return evalUnaryOperator(ctx, ops.UnaryBitwiseNotOp.Caption(), a.Next)
	default:
		return a.Exponent.Evaluate(ctx, valueToSet)
	}
}

func (a *Exponent) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	return evalBinaryOperator(ctx, valueToSet, ops.PowOp.Caption(), a.Primary, a.Next)
}

func (a *Primary) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	if a.SubExpression != nil {
		if valueToSet != nil {
			// TODO: change to normal description
			return nil, errors.New("unable to set to subexpression evaluation")
		}

		return a.SubExpression.Evaluate(ctx, valueToSet)
	}

	if a.Constant != nil {
		if valueToSet != nil {
			// TODO: change to normal description
			return nil, errors.New("unable to set to constant")
		}

		return a.Constant.Evaluate(ctx)
	}

	if a.RandomAccess != nil {
		return a.RandomAccess.Evaluate(ctx, valueToSet)
	}

	if a.AttributeAccess != nil {
		return a.AttributeAccess.Evaluate(ctx, valueToSet, nil)
	}

	if a.LambdaDef != nil {
		return a.LambdaDef.Evaluate(ctx)
	}

	panic("unreachable")
}

func (c *Constant) Evaluate(ctx common.Context) (common.Type, error) {
	if c.Integer != nil {
		return types.NewIntegerInstance(*c.Integer), nil
	}

	if c.Real != nil {
		return types.NewRealInstance(*c.Real), nil
	}

	if c.Bool != nil {
		return types.NewBoolInstance(bool(*c.Bool)), nil
	}

	if c.String != nil {
		return types.NewStringInstance(*c.String), nil
	}

	if c.List != nil {
		list := types.NewListInstance()
		for _, expr := range c.List {
			value, err := expr.Evaluate(ctx, nil)
			if err != nil {
				return nil, err
			}

			list.Values = append(list.Values, value)
		}

		return list, nil
	}

	if c.Dictionary != nil {
		dict := types.NewDictionaryInstance()
		for _, entry := range c.Dictionary {
			key, value, err := entry.Evaluate(ctx)
			if err != nil {
				return nil, err
			}

			if err := dict.SetElement(key, value); err != nil {
				return nil, err
			}
		}

		return dict, nil
	}

	panic("unreachable")
}

func (d *DictionaryEntry) Evaluate(ctx common.Context) (common.Type, common.Type, error) {
	key, err := d.Key.Evaluate(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	value, err := d.Value.Evaluate(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	return key, value, nil
}

func (a *AttributeAccess) Evaluate(ctx common.Context, valueToSet, prevValue common.Type) (common.Type, error) {
	if valueToSet != nil {
		// set
		var currentValue common.Type
		var err error = nil
		if a.Ident != nil {
			if *a.Ident == "нуль" {
				return nil, util.RuntimeError("неможливо встановити значення об'єкту 'нуль'")
			}

			if a.AttributeAccess != nil {
				currentValue, err = getCurrentValue(ctx, prevValue, *a.Ident)
			} else {
				currentValue = valueToSet
			}
		} else if a.CallFunc != nil {
			if a.AttributeAccess == nil {
				return nil, util.RuntimeError("неможливо присвоїти значення виклику функції")
			}

			function, err := getCurrentValue(ctx, prevValue, a.CallFunc.Ident)
			if err != nil {
				return nil, err
			}

			currentValue, err = a.CallFunc.Evaluate(ctx, function)
			if err != nil {
				return nil, err
			}
		}

		if a.AttributeAccess != nil {
			currentValue, err = a.AttributeAccess.Evaluate(ctx, valueToSet, currentValue)
			if err != nil {
				return nil, err
			}
		}

		if prevValue != nil {
			err = nil
			if a.Ident != nil {
				_, err = prevValue.SetAttribute(*a.Ident, currentValue)
			} else if a.CallFunc != nil {
				// ignore
			}

			if err != nil {
				return nil, err
			}

			return prevValue, nil
		}

		if a.Ident != nil {
			return currentValue, ctx.SetVar(*a.Ident, currentValue)
		}

		return currentValue, nil
	} else {
		// get
		var currentValue common.Type
		var err error = nil
		if a.Ident != nil {
			if *a.Ident == "нуль" {
				if prevValue == nil {
					return types.NewNilInstance(), nil
				} else {
					return nil, util.RuntimeError("'нуль' не є атрибутом")
				}
			}

			currentValue, err = getCurrentValue(ctx, prevValue, *a.Ident)
			if err != nil {
				return nil, err
			}
		} else if a.CallFunc != nil {
			variable, err := getCurrentValue(ctx, prevValue, a.CallFunc.Ident)
			if err != nil {
				return nil, err
			}

			currentValue, err = a.CallFunc.Evaluate(ctx, variable)
			if err != nil {
				return nil, errors.New(
					fmt.Sprintf(
						"  Файл \"%s\", рядок %d, позиція %d\n    %s\n%s",
						a.CallFunc.Pos.Filename, a.CallFunc.Pos.Line, a.CallFunc.Pos.Column, "TODO", err.Error(),
					),
				)
			}
		}

		if a.AttributeAccess != nil {
			return a.AttributeAccess.Evaluate(ctx, valueToSet, currentValue)
		}

		return currentValue, err
	}
}

func (a *RandomAccess) Evaluate(ctx common.Context, valueToSet common.Type) (common.Type, error) {
	variable, err := ctx.GetVar(a.Ident)
	if err != nil {
		return nil, err
	}

	if valueToSet != nil {
		variable, err = evalSingleSetByIndexOperation(ctx, variable, a.Index, valueToSet)
		if err != nil {
			return nil, err
		}

		return variable, ctx.SetVar(a.Ident, variable)
	}

	for _, indexExpression := range a.Index {
		index, err := indexExpression.Evaluate(ctx, nil)
		if err != nil {
			return nil, err
		}

		variable, err = evalSingleGetByIndexOperation(variable, index)
		if err != nil {
			return nil, err
		}
	}

	return variable, nil
}

func (l *LambdaDef) Evaluate(ctx common.Context) (common.Type, error) {
	arguments := evalParameters(ctx, l.Parameters)
	returnTypes := evalReturnTypes(ctx, l.ReturnTypes)
	return types.NewFunctionInstance(
		"",
		arguments,
		func(context interface{}, _ *[]common.Type, kwargs *map[string]common.Type) (common.Type, error) {
			funcContext := context.(common.FunctionContext)
			return l.Body.Evaluate(funcContext.Context)
		},
		returnTypes,
		false,
		ctx.GetPackage().(*types.PackageInstance),
		"", // TODO: add doc
	), nil
}

func (a *CallFunc) Evaluate(ctx common.Context, variable common.Type) (common.Type, error) {
	switch function := variable.(type) {
	case *types.Class:
		callable, err := function.GetAttribute(ops.ConstructorName)
		if err != nil {
			return nil, err
		}

		switch __constructor__ := callable.(type) {
		case *types.FunctionInstance:
			instance, err := function.GetEmptyInstance()
			if err != nil {
				return nil, err
			}

			args := []common.Type{instance}
			kwargs := map[string]common.Type{__constructor__.Arguments[0].Name: instance}

			// TODO: check if constructor returns nothing.
			_, err = a.evalFunction(ctx, __constructor__, &args, &kwargs, 1)
			if err != nil {
				return nil, err
			}

			// gotVariadic := false
			// for i, expressionArgument := range a.Arguments {
			// 	arg, err := expressionArgument.Evaluate(ctx, nil)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			//
			// 	args = append(args, arg)
			// 	if !gotVariadic {
			// 		gotVariadic = constructor.Arguments[i+1].IsVariadic
			// 		kwargs[constructor.Arguments[i+1].Name] = arg
			// 	}
			// }
			//
			// if err := types.CheckFunctionArguments(constructor, &args, &kwargs); err != nil {
			// 	return nil, err
			// }
			//
			// ctx.PushScope(kwargs)
			//
			// // TODO: check if constructor returns nothing.
			// // TODO: check if constructor returns nothing.
			// _, err = constructor.Call(nil, &args, &kwargs)
			// if err != nil {
			// 	return nil, err
			// }
			//
			// ctx.PopScope()
			return args[0], nil
		default:
			return nil, util.ObjectIsNotCallable(a.Ident, callable.GetTypeName())
		}
	case *types.FunctionInstance:
		var args []common.Type
		kwargs := map[string]common.Type{}
		return a.evalFunction(ctx, function, &args, &kwargs, 0)
		// gotVariadic := false
		// for i, expressionArgument := range a.Arguments {
		// 	arg, err := expressionArgument.Evaluate(ctx, nil)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		//
		// 	args = append(args, arg)
		// 	if !gotVariadic {
		// 		gotVariadic = object.Arguments[i].IsVariadic
		// 		kwargs[object.Arguments[i].Name] = arg
		// 	}
		// }
		//
		// if err := types.CheckFunctionArguments(object, &args, &kwargs); err != nil {
		// 	return nil, err
		// }
		//
		// ctx.PushScope(kwargs)
		// res, err := object.Call(ParserInstance, &args, &kwargs)
		// if err != nil {
		// 	return nil, err
		// }
		//
		// if err := types.CheckResult(res, object); err != nil {
		// 	return nil, err
		// }
		//
		// ctx.PopScope()
		// return res, nil
	case types.Instance:
		operator, err := function.GetClass().GetAttribute(ops.CallOperatorName)
		if err != nil {
			return nil, err
		}

		switch __call__ := operator.(type) {
		case *types.FunctionInstance:
			args := []common.Type{variable}
			kwargs := map[string]common.Type{__call__.Arguments[0].Name: variable}
			return a.evalFunction(ctx, __call__, &args, &kwargs, 1)
			// gotVariadic := false
			// for i, expressionArgument := range a.Arguments {
			// 	arg, err := expressionArgument.Evaluate(ctx, nil)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			//
			// 	args = append(args, arg)
			// 	if !gotVariadic {
			// 		gotVariadic = callOperator.Arguments[i+1].IsVariadic
			// 		kwargs[callOperator.Arguments[i+1].Name] = arg
			// 	}
			//
			// }
			//
			// if err := types.CheckFunctionArguments(callOperator, &args, &kwargs); err != nil {
			// 	return nil, err
			// }
			//
			// ctx.PushScope(kwargs)
			// res, err := callOperator.Call(nil, &args, &kwargs)
			// if err != nil {
			// 	return nil, err
			// }
			//
			// if err := types.CheckResult(res, callOperator); err != nil {
			// 	return nil, err
			// }
			//
			// ctx.PushScope(kwargs)
			// return res, nil
		default:
			return nil, util.ObjectIsNotCallable(a.Ident, operator.GetTypeName())
		}
	default:
		return nil, util.ObjectIsNotCallable(a.Ident, function.GetTypeName())
	}
}

func (a *CallFunc) evalFunction(
	ctx common.Context,
	function *types.FunctionInstance,
	args *[]common.Type,
	kwargs *map[string]common.Type,
	argsShift int,
) (common.Type, error) {
	variadicArgs := types.NewListInstance()
	variadicArgsIndex := -1
	for i, expressionArgument := range a.Arguments {
		arg, err := expressionArgument.Evaluate(ctx, nil)
		if err != nil {
			return nil, err
		}

		*args = append(*args, arg)
		if variadicArgsIndex == -1 {
			if function.Arguments[i+argsShift].IsVariadic {
				variadicArgsIndex = i + argsShift
				variadicArgs.Values = append(variadicArgs.Values, arg)
			} else {
				(*kwargs)[function.Arguments[i+argsShift].Name] = arg
			}
		} else {
			variadicArgs.Values = append(variadicArgs.Values, arg)
		}
	}

	if variadicArgsIndex != -1 {
		(*kwargs)[function.Arguments[variadicArgsIndex].Name] = variadicArgs
	}

	if err := types.CheckFunctionArguments(function, args, kwargs); err != nil {
		return nil, err
	}

	ctx.PushScope(*kwargs)
	funcContext := common.FunctionContext{
		Context: ctx,
		Parser:  ParserInstance,
	}
	res, err := function.Call(funcContext, args, kwargs)
	if err != nil {
		return nil, err
	}

	if err := types.CheckResult(res, function); err != nil {
		return nil, err
	}

	ctx.PopScope()
	return res, nil
}

func getCurrentValue(ctx common.Context, prevValue common.Type, identifier string) (common.Type, error) {
	if prevValue != nil {
		return prevValue.GetAttribute(identifier)
	} else {
		return ctx.GetVar(identifier)
	}
}
