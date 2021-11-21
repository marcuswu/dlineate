package dlineate

func CoincidentConstraint(p1 Point, p2 Point) *Constraint {
	constraint := emptyConstraint()
	append(constraint.elements, p1)
	append(constraint.elements, p2)

	return constraint
}