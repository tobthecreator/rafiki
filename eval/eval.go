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

	// Base case, top node
	case *ast.Program:
		return evalStatements(node.Statements)

	// Second layer base case, individual statements
	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)

		return evalInfixExpression(node.Operator, left, right)

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

func evalStatements(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)
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
		return NULL
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
	default:
		return NULL
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
		return NULL

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
		return NULL
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
