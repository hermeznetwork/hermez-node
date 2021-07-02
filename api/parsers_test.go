package api

type queryParser struct {
	m map[string]string
}

func (qp *queryParser) Query(query string) string {
	if val, ok := qp.m[query]; ok {
		return val
	}
	return ""
}

func (qp *queryParser) Param(param string) string {
	if val, ok := qp.m[param]; ok {
		return val
	}
	return ""
}
