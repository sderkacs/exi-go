package core

import (
	"strings"

	"github.com/sderkacs/go-exi/utils"
)

func QNameCompareFunc(q1, q2 utils.QName) int {
	return QNameCompare(q1.Space, q1.Local, q2.Space, q2.Local)
}

func QNameCompare(ns1, ln1, ns2, ln2 string) int {
	clp := strings.Compare(ln1, ln2)
	if clp == 0 {
		return strings.Compare(ns1, ns2)
	} else {
		return clp
	}
}

func AttributeCompareFunc(a1, a2 *Attribute) int {
	q1 := a1.GetQName()
	q2 := a2.GetQName()

	return QNameCompare(q1.Space, q1.Local, q2.Space, q2.Local)
}
