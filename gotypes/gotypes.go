package gotypes

import (
	"errors"
	"go/token"
	"go/types"
)

func ParseSignature(expr string) (*types.Signature, error) {
	tv, err := types.Eval(token.NewFileSet(), nil, token.NoPos, expr)
	if err != nil {
		return nil, err
	}
	if tv.Value != nil {
		return nil, errors.New("signature expression should have nil value")
	}
	s, ok := tv.Type.(*types.Signature)
	if !ok {
		return nil, errors.New("provided type is not a function signature")
	}
	return s, nil
}
