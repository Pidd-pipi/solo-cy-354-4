// Package constants centralizes business constants for the campus-market backend.
package constants

// ProductCategory defines product category enum values shared with the frontend.
const (
	ProductCategoryBooks    = "books"
	ProductCategoryElectronics = "electronics"
	ProductCategoryDaily    = "daily"
	ProductCategoryClothing = "clothing"
)

// ProductCategories lists all valid product categories.
var ProductCategories = []string{
	ProductCategoryBooks, ProductCategoryElectronics, ProductCategoryDaily, ProductCategoryClothing,
}

// IsProductCategory reports whether the given category is valid.
func IsProductCategory(c string) bool {
	for _, v := range ProductCategories {
		if v == c {
			return true
		}
	}
	return false
}

// ProductStatus defines product lifecycle states shared with the frontend.
const (
	ProductStatusOnSale  = "on_sale"
	ProductStatusReserved = "reserved"
	ProductStatusSold    = "sold"
	ProductStatusRemoved = "removed"
)

// ProductStatuses lists all valid product statuses.
var ProductStatuses = []string{
	ProductStatusOnSale, ProductStatusReserved, ProductStatusSold, ProductStatusRemoved,
}

// IsProductStatus reports whether the given status is valid.
func IsProductStatus(s string) bool {
	for _, v := range ProductStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// ProductCategoryText returns the Chinese label of a product category.
func ProductCategoryText(c string) string {
	switch c {
	case ProductCategoryBooks:
		return "书籍"
	case ProductCategoryElectronics:
		return "电子产品"
	case ProductCategoryDaily:
		return "生活用品"
	case ProductCategoryClothing:
		return "服饰"
	default:
		return "未知"
	}
}

// ProductStatusText returns the Chinese label of a product status.
func ProductStatusText(s string) string {
	switch s {
	case ProductStatusOnSale:
		return "在售"
	case ProductStatusReserved:
		return "已预订"
	case ProductStatusSold:
		return "已售出"
	case ProductStatusRemoved:
		return "已下架"
	default:
		return "未知"
	}
}
