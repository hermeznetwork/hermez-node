package parsers

// Pagination type for holding pagination params
type Pagination struct {
	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}
