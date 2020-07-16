package statements

func (statement *Statement) IsNilIncluded(requiredField, includeNil bool) bool {
	return requiredField || includeNil
}
