package models

type BrandPrimaryKey struct {
	Id string `json:"id"`
}

type CreateBrand struct {
	Name string `json:"name"`
}

type Brand struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateBrand struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetListBrandRequest struct {
	Offset int64  `json:"offset"`
	Limit  int64  `json:"limit"`
	Search string `json:"search"`
	Query  string `json:"query"`
}

type GetListBrandResponse struct {
	Count   int      `json:"count"`
	Brands []*Brand `json:"brands"`
}
