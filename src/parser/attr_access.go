package parser

import (
	"errors"
	"github.com/YuriyLisovskiy/borsch/src/ast"
	"github.com/YuriyLisovskiy/borsch/src/models"
)

func (p *Parser) parseAttrAccess(parent ast.ExpressionNode) (ast.ExpressionNode, error) {
	if name := p.match(models.TokenTypesList[models.Name]); name != nil {
		var attr ast.ExpressionNode = nil
		if p.match(models.TokenTypesList[models.LPar]) != nil {
			var err error
			attr, err = p.parseFunctionCall(name, parent)
			if err != nil {
				return nil, err
			}

			attr, err = p.parseRandomAccessOperation(attr)
			if err != nil {
				return nil, err
			}

			//if randomAccessOp != nil {
			//	parent = randomAccessOp
			//}
			//name = nil
		} else {
			attr = ast.NewVariableNode(*name)
		}

		var baseToken *models.Token = nil
		switch node := parent.(type) {
		case ast.AttrOpNode:
			//baseToken = node.Attr
		case ast.VariableNode:
			baseToken = &node.Variable
		}

		parent = ast.NewGetAttrOpNode(baseToken, parent, attr, name.Row)
		randomAccessOp, err := p.parseRandomAccessOperation(parent)
		if err != nil {
			return nil, err
		}

		if randomAccessOp != nil {
			parent = randomAccessOp
		}

		if dot := p.match(models.TokenTypesList[models.AttrAccessOp]); dot != nil {
			return p.parseAttrAccess(parent)
		}

		assignOperator := p.match(models.TokenTypesList[models.Assign])
		if assignOperator != nil {
			rightNode, err := p.parseFormula()
			if err != nil {
				return nil, err
			}

			binaryNode := ast.NewBinOperationNode(*assignOperator, parent, rightNode)
			return binaryNode, nil
		}

		return parent, nil
	}

	return nil, errors.New("очікується назва атрибута")
}
