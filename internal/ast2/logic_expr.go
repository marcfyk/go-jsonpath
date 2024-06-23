package ast2

type LogicExprNot struct {
	LogicExpr LogicExpr
}

func (l LogicExprNot) Evaluate(n Node) bool {
	return !l.LogicExpr.Evaluate(n)
}

type LogicExprOr struct {
	LogicExprs []LogicExpr
}

func (l LogicExprOr) Evaluate(n Node) bool {
	for _, logicExpr := range l.LogicExprs {
		if logicExpr.Evaluate(n) {
			return true
		}
	}
	return false
}

type LogicExprAnd struct {
	LogicExprs []LogicExpr
}

func (l LogicExprAnd) Evaluate(n Node) bool {
	for _, logicExpr := range l.LogicExprs {
		if !logicExpr.Evaluate(n) {
			return false
		}
	}
	return true
}
