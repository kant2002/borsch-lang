package builtin

import (
	"errors"
	"fmt"
	"github.com/YuriyLisovskiy/borsch/src/util"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	noneType = iota
	realNumberType
	integerNumberType
	stringType
	boolType
)

type ValueType interface {
	String() string
	Representation() string
	TypeHash() int
	TypeName() string
}

type IterableType interface {
	Length() int64
	GetElement(int64) (ValueType, error)
	SetElement(int64, ValueType) (ValueType, error)
}

func getIndex(index, length int64) (int64, error) {
	if index >= 0 && index < length {
		return index, nil
	} else if index < 0 && index >= -length {
		return length + index, nil
	}

	return 0, errors.New("індекс рядка за межами послідовності")
}

// NoneType represents none type.
type NoneType struct {
}

func (t NoneType) String() string {
	return "NoneType{" + t.Representation() + "}"
}

func (t NoneType) Representation() string {
	return "Порожнеча"
}

func (t NoneType) TypeHash() int {
	return noneType
}

func (t NoneType) TypeName() string {
	return "ніякий"
}

// RealNumberType represents numbers as float64
type RealNumberType struct {
	Value float64
}

func NewRealNumberType(value string) (RealNumberType, error) {
	number, err := strconv.ParseFloat(strings.TrimSuffix(value, "f"), 64)
	if err != nil {
		return RealNumberType{}, util.RuntimeError(err.Error())
	}

	return RealNumberType{Value: number}, nil
}

func (t RealNumberType) String() string {
	return "RealType{" + t.Representation() + "}"
}

func (t RealNumberType) Representation() string {
	return fmt.Sprintf("%f", t.Value)
}

func (t RealNumberType) TypeHash() int {
	return realNumberType
}

func (t RealNumberType) TypeName() string {
	return "дійсне"
}

// IntegerNumberType represents numbers as float64
type IntegerNumberType struct {
	Value int64
}

func NewIntegerNumberType(value string) (IntegerNumberType, error) {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return IntegerNumberType{}, util.RuntimeError(err.Error())
	}

	return IntegerNumberType{Value: number}, nil
}

func (t IntegerNumberType) String() string {
	return "IntegerType{" + t.Representation() + "}"
}

func (t IntegerNumberType) Representation() string {
	return fmt.Sprintf("%d", t.Value)
}

func (t IntegerNumberType) TypeHash() int {
	return integerNumberType
}

func (t IntegerNumberType) TypeName() string {
	return "ціле"
}

// StringType is string representation.
type StringType struct {
	Value string
}

func (t StringType) String() string {
	return "StringType{\"" + t.Representation() + "\"}"
}

func (t StringType) Representation() string {
	return t.Value
}

func (t StringType) TypeHash() int {
	return stringType
}

func (t StringType) TypeName() string {
	return "рядок"
}

func (t StringType) Length() int64 {
	return int64(utf8.RuneCountInString(t.Value))
}

func (t StringType) GetElement(index int64) (ValueType, error) {
	idx, err := getIndex(index, t.Length())
	if err != nil {
		return nil, err
	}

	return StringType{Value: string([]rune(t.Value)[idx])}, nil
}

func (t StringType) SetElement(index int64, value ValueType) (ValueType, error) {
	idx, err := getIndex(index, t.Length())
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case StringType:
		if utf8.RuneCountInString(v.Value) != 1 {
			return nil, errors.New("неможливо вставити жодного, або більше ніж один символ в рядок")
		}

		runes := []rune(v.Value)
		target := []rune(t.Value)
		target[idx] = runes[0]
		t.Value = string(target)
	default:
		return nil, errors.New(fmt.Sprintf("неможливо вставити в рядок об'єкт типу '%s'", v.TypeName()))
	}

	return t, nil
}

// BoolType is string representation.
type BoolType struct {
	Value bool
}

func NewBoolType(value string) (BoolType, error) {
	switch value {
	case "істина":
		value = "t"
	case "хиба":
		value = "f"
	}

	boolean, err := strconv.ParseBool(value)
	if err != nil {
		return BoolType{}, util.RuntimeError(err.Error())
	}

	return BoolType{Value: boolean}, nil
}

func (t BoolType) String() string {
	return "\"" + t.Representation() + "\""
}

func (t BoolType) Representation() string {
	if t.Value {
		return "істина"
	}

	return "хиба"
}

func (t BoolType) TypeHash() int {
	return boolType
}

func (t BoolType) TypeName() string {
	return "логічне"
}
