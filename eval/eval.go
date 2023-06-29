package eval

import (
	"fmt"
	"rafiki/ast"
	"rafiki/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Call recusively while swinging through the tree
func Eval(node ast.Node) object.Object {
	switch node := node.(type) {

	// Base case, top node of the program or a top node of a block
	case *ast.Program:
		return evalProgram(node.Statements)

	case *ast.BlockStatement:
		return evalBlockStatement(node)

	// Second layer base case, individual statements
	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	case *ast.PrefixExpression:
		right := Eval(node.Right)

		if isError(right) {
			return right
		}

		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left)

		if isError(left) {
			return left
		}

		right := Eval(node.Right)

		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)

		if isError(val) {
			return val
		}

		return &object.ReturnValue{Value: val}

	// Leaves, the objects
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	default:
		fmt.Printf("node: %v\n", node)
	}

	return NULL
}

func evalProgram(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

/*
Functionally similar to evalProgram(), but requires a separate implementation
to support return statements that may be nested deeply in the block statement

# In the case

x = 10

	if (x > 5) {
		if (x > 9) {
			return 999
		}

		return 1
	}

evalProgram returns 1.

evalBlockStatement() returns as soon as it encounters a return statement
*/
func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement)

		if result != nil {
			rt := result.Type()

			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalNegatePrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s %s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == rightType && leftType == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())

	}
}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}

	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())

	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case NULL:
		return TRUE
	case FALSE:
		return TRUE
	case TRUE:
		return FALSE
	default:
		return FALSE
	}
}

func evalNegatePrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func evalIfExpression(expr *ast.IfExpression) object.Object {
	condition := Eval(expr.Condition)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(expr.Consequence)
	}

	// If else {}, else
	if expr.Alternative != nil {
		return Eval(expr.Alternative)
	}

	return NULL
}

func isTruthy(o object.Object) bool {
	switch o {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(o object.Object) bool {
	if o != nil {
		return o.Type() == object.ERROR_OBJ
	}

	return false
}
