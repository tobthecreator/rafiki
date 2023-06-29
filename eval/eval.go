package eval

import (
	"fmt"
	"rafiki/ast"
	"rafiki/object"
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

	// Leaves, the objects
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	default:
		fmt.Printf("node: %v\n", node)
	}

	return nil
}

func evalStatements(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)
	}

	return result
}
