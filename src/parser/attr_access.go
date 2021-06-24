package parser

import (
	"errors"
	"github.com/YuriyLisovskiy/borsch/src/ast"
	"github.com/YuriyLisovskiy/borsch/src/models"
)

func (p *Parser) parseAttrAccess(base ast.ExpressionNode) (ast.ExpressionNode, error) {
	if name := p.match(models.TokenTypesList[models.Name]); name != nil {
		var attr ast.ExpressionNode = nil
		if p.match(models.TokenTypesList[models.LPar]) != nil {
			var err error
			attr, err = p.parseFunctionCall(name)
			if err != nil {
				return nil, err
			}
		} else {
			attr = ast.NewVariableNode(*name)
		}

		result, err := p.parseRandomAccessOperation(ast.NewGetAttrOpNode(base, attr, name.Row))
		if err != nil {
			return nil, err
		}

		if dot := p.match(models.TokenTypesList[models.AttrAccessOp]); dot != nil {
			return p.parseAttrAccess(result)
		}

		assignOperator := p.match(models.TokenTypesList[models.Assign])
		if assignOperator != nil {
			rightNode, err := p.parseFormula()
			if err != nil {
				return nil, err
			}

			binaryNode := ast.NewBinOperationNode(*assignOperator, result, rightNode)
			return binaryNode, nil
		}

		return result, nil
	}

	return nil, errors.New("очікується назва атрибута")
}
