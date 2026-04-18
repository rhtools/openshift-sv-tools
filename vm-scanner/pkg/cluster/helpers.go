package cluster

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

func isAPIMissing(err error) bool {
	if apierrors.IsNotFound(err) {
		return true
	}
	return meta.IsNoMatchError(err)
}
