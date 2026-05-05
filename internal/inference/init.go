package inference

import "structlens/internal/model"

func init() {
	model.SetTypeResolver(ResolveType)
}
